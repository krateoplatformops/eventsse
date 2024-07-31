package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	serviceName := "TestService"
	var healthy int32

	// Configura il handler con lo stato di salute
	handler := Check(&healthy, serviceName)

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/health", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 Method Not Allowed, got %v", rr.Code)
		}
		if rr.Header().Get("Allow") != "GET" {
			t.Errorf("expected Allow header to be 'GET', got %v", rr.Header().Get("Allow"))
		}
	})

	t.Run("Service Healthy", func(t *testing.T) {
		atomic.StoreInt32(&healthy, 1)
		req, err := http.NewRequest(http.MethodGet, "/health", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %v", rr.Code)
		}
		if rr.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type header to be 'application/json', got %v", rr.Header().Get("Content-Type"))
		}

		var response map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("could not parse response body: %v", err)
		}
		if response["name"] != serviceName {
			t.Errorf("expected service name %q, got %q", serviceName, response["name"])
		}
	})

	t.Run("Service Unhealthy", func(t *testing.T) {
		atomic.StoreInt32(&healthy, 0)
		req, err := http.NewRequest(http.MethodGet, "/health", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503 Service Unavailable, got %v", rr.Code)
		}
	})
}
