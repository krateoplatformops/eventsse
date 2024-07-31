package subscriber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krateoplatformops/eventsse/internal/cache"
	"github.com/krateoplatformops/eventsse/internal/labels"
	"github.com/krateoplatformops/eventsse/internal/store"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ store.Store = (*MockStore)(nil)

// MockStore è un mock del client store per testare l'handler
type MockStore struct {
	data map[string]corev1.Event
}

func (m *MockStore) PrepareKey(uid, compositionID string) string {
	return uid + ":" + compositionID
}

func (m *MockStore) Set(key string, event *corev1.Event) error {
	if m.data == nil {
		m.data = make(map[string]corev1.Event)
	}
	m.data[key] = *event
	return nil
}

func (m *MockStore) Get(key string, opts store.GetOptions) (data []corev1.Event, found bool, err error) {
	event, exists := m.data[key]
	if !exists {
		return nil, false, fmt.Errorf("key '%s' not found", key)
	}
	return []corev1.Event{event}, true, nil
}

func (m *MockStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStore) SetTTL(_ int) {

}

func (m *MockStore) Close() error {
	return nil
}

func TestServeHTTP(t *testing.T) {
	ttlCache := cache.NewTTL[string, corev1.Event]()
	ms := &MockStore{}

	handler := Handle(HandleOptions{TTLCache: ttlCache, Store: ms})

	t.Run("Malformed JSON", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/events", bytes.NewBuffer([]byte("{malformed json")))
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Empty Body", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(nil))
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status 204 No Content, got %v", rr.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		event := corev1.Event{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-event",
				Namespace: "demo-system",
				UID:       types.UID("test-uid"),
			},
			Message: "Test Event",
		}

		eventBytes, _ := json.Marshal(event)
		req, err := http.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(eventBytes))
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %v", rr.Code)
		}

		expectedKey := ms.PrepareKey(string(event.UID), labels.CompositionID(&event))
		if rr.Body.String() != expectedKey {
			t.Errorf("expected response body %q, got %q", expectedKey, rr.Body.String())
		}
	})
}
