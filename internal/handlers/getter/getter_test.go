package getter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krateoplatformops/eventsse/internal/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ store.Store = (*MockStore)(nil)

// MockStore è un mock del client store per testare l'handler
type MockStore struct {
	data map[string]corev1.Event
}

func (m *MockStore) PrepareKey(uid, compositionID string) string {
	return compositionID
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
		return nil, false, nil
	}
	return []corev1.Event{event}, true, nil
}

func (m *MockStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStore) SetTTL(_ int) {}

func (m *MockStore) Close() error {
	return nil
}

func TestEventsHandler(t *testing.T) {
	handler := Events(&MockStore{
		data: map[string]corev1.Event{
			"comp1": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-event-2",
					Namespace: "demo-system",
					UID:       types.UID("evt1"),
				},
				Message: "Test Event 1",
			},
			"comp2": {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-event-2",
					Namespace: "demo-system",
					UID:       types.UID("evt2"),
				},
				Message: "Test Event 2",
			},
		},
	}, 10)

	t.Run("Valid request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/events?composition=comp1", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %v", rr.Code)
		}

		var events []corev1.Event
		if err := json.NewDecoder(rr.Body).Decode(&events); err != nil {
			t.Errorf("could not decode response: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("No events found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/events?composition=comp3", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status 204 No Content, got %v", rr.Code)
		}
	})
}
