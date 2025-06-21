package service

import (
	"compress/gzip"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	countShortKey = 8
)

func GenerateShortKey() string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, countShortKey)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}
		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if !acceptsGzip {
			next.ServeHTTP(w, r)
			return
		}
		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()
		w.Header().Set("Content-Encoding", "gzip")
		gzw := gzipWriter{
			ResponseWriter: w,
			Writer:         gzWriter,
		}
		next.ServeHTTP(gzw, r)
	})

}
