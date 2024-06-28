package handlers

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type Handler interface {
	Name() string
	Pattern() string
	Methods() []string
	Handler() http.HandlerFunc
}

func Serve(all []Handler) http.Handler {
	return http.HandlerFunc(func(wri http.ResponseWriter, req *http.Request) {
		var allow []string
		for _, route := range all {
			if req.URL.Path != route.Pattern() {
				continue
			}

			if !contains(route.Methods(), req.Method) {
				allow = append(allow, route.Methods()...)
				continue
			}

			l := log.With().Logger()
			req = req.WithContext(l.WithContext(req.Context()))
			route.Handler()(wri, req)
			return
		}

		if len(allow) > 0 {
			wri.Header().Set("Allow", strings.Join(allow, ", "))
			http.Error(wri, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}

		http.NotFound(wri, req)
	})
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}
