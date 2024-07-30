package health

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

func Check(healthy *int32, serviceName string) http.Handler {
	return &handler{
		healthy:     healthy,
		serviceName: serviceName,
	}
}

var _ http.Handler = (*handler)(nil)

type handler struct {
	healthy     *int32
	serviceName string
}

// @title EventSSE API
// @version 1.0
// @description This the Krateo EventSSE server.
// @BasePath /

// Health godoc
// @Summary Liveness Endpoint
// @Description Health Check
// @ID health
// @Produce  json
// @Success 200 {object} map[string]any
// @Router /health [get]
func (r *handler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		wri.Header().Set("Allow", "GET")
		http.Error(wri, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if atomic.LoadInt32(r.healthy) == 1 {
		data := map[string]string{
			"name": r.serviceName,
			//"version": r.version,
		}

		wri.Header().Set("Content-Type", "application/json")
		wri.WriteHeader(http.StatusOK)
		json.NewEncoder(wri).Encode(data)
		return
	}
	wri.WriteHeader(http.StatusServiceUnavailable)
}
