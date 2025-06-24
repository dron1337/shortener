package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dron1337/shortener/internal/contextkeys"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
)

type ContextKey string

var (
	hashKey  = securecookie.GenerateRandomKey(64)
	blockKey = securecookie.GenerateRandomKey(32)
	s        = securecookie.New(hashKey, blockKey)
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		// 1. Пытаемся получить и декодировать куку
		if cookie, err := r.Cookie("session"); err == nil {
			var sessionData map[string]string
			if err := s.Decode("session", cookie.Value, &sessionData); err == nil {
				if id, exists := sessionData["user_id"]; exists {
					userID = id
				}
			}
		}

		// 2. Если куки нет или она невалидна - создаем новую
		if userID == "" {
			userID = generateUserID() // Ваша функция генерации ID
			value := map[string]string{"user_id": userID}
			if encoded, err := s.Encode("session", value); err == nil {
				cookie := &http.Cookie{
					Name:     "session",
					Value:    encoded,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   86400, // 1 день
				}
				http.SetCookie(w, cookie)
			} else {
				fmt.Printf("Ошибка кодирования куки: %v", err)
			}
		}

		// 3. Добавляем userID в контекст запроса
		ctx := context.WithValue(r.Context(), contextkeys.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateUserID() string {
	return uuid.New().String()
}
