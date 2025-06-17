package service

import (
	"compress/gzip"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

const (
	countShortKey = 8
)

func GenerateShortKey(r *rand.Rand) string {
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
		contentTypes := []string{
			"application/javascript",
			"application/json",
			"text/css",
			"text/html",
			//"text/plain",
			"text/xml",
		}
		log.Println("Content-Type=", r.Header.Get("Content-Type"))
		log.Println("Accept-Encoding=", r.Header.Get("Accept-Encoding"))
		if !containsContentType(r.Header.Get("Content-Type"), contentTypes) {
			log.Println("next.ServeHTT")
			next.ServeHTTP(w, r)
			return
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			log.Println("next.ServeHTT")
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			log.Println("Error NewWriterLevel", err)
			return
		}
		defer gz.Close()
		log.Printf("Set Content-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func containsContentType(contentType string, types []string) bool {
	for _, t := range types {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}
