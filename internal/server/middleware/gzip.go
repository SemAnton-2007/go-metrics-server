package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// GzipMiddleware creates a middleware that handles gzip compression.
// It decompresses incoming requests with Content-Encoding: gzip header
// and compresses responses when client accepts gzip encoding.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Распаковка входящего запроса
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip encoding", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		// Проверяем, нужно ли сжимать ответ
		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if !acceptsGzip {
			next.ServeHTTP(w, r)
			return
		}

		// Устанавливаем заголовки
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Del("Content-Length")

		// Получаем gzip writer из пула
		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			gz.Close()
			gzipWriterPool.Put(gz)
		}()

		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
