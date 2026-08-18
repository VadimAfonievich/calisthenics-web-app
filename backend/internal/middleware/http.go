package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (writer *responseWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *responseWriter) Write(body []byte) (int, error) {
	if writer.status == 0 {
		writer.status = http.StatusOK
	}
	return writer.ResponseWriter.Write(body)
}

func WithRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		requestID := request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = startedAt.UTC().Format("20060102T150405.000000000Z")
		}
		writer.Header().Set("X-Request-ID", requestID)
		response := &responseWriter{ResponseWriter: writer}
		next.ServeHTTP(response, request)
		logger.Info("request completed", "request_id", requestID, "method", request.Method, "path", request.URL.Path, "status", response.status, "duration_ms", time.Since(startedAt).Milliseconds())
	})
}
