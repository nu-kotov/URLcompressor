package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nu-kotov/URLcompressor/internal/app/compress"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	logFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Log.Info("request",
			zap.String("method", r.Method),
			zap.String("uri", r.URL.Path),
			zap.String("duration", duration.String()),
		)
		logger.Log.Info("response",
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	})
	return http.HandlerFunc(logFn)
}

func RequestCompressor(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := compress.NewCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			fmt.Println("Тело", r.Body)
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	}
}
