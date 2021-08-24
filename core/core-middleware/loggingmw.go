package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware returns a new instance of the logging middleware
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// LoggingMiddleware runs the logging middleware
func (m *LoggingMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.Info(
			"request started",
			zap.Any("proto", r.Proto),
			zap.Any("uri", r.RequestURI),
				zap.Any("method", r.Method),
					zap.Any("remote", r.RemoteAddr),
						zap.Any("user-agent", r.UserAgent()),
		)
		next.ServeHTTP(w, r)
	})
}
