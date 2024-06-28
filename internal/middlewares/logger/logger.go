package routes

import (
	"net/http"

	"github.com/rs/zerolog"
)

// Logger returns a zero allocation JSON logger middleware.
func Logger(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := log.With().Logger()

			// l.WithContext returns a copy of the context with the log object associated
			r = r.WithContext(l.WithContext(r.Context()))

			next.ServeHTTP(w, r)
		})
	}
}
