package publisher

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/krateoplatformops/eventsse/internal/cache"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServeHTTP(t *testing.T) {

	t.Run("Send events from cache", func(t *testing.T) {
		const exp = `event: krateo
id: event1
data: {"metadata":{"name":"event1","namespace":"demo-system","creationTimestamp":null},"involvedObject":{},"source":{},"firstTimestamp":null,"lastTimestamp":null,"eventTime":null,"reportingComponent":"","reportingInstance":""}

`

		ttlCache := cache.NewTTL[string, corev1.Event]()
		defer func() {
			ttlCache.Clear()
		}()
		ttlCache.Set("event1", corev1.Event{
			ObjectMeta: v1.ObjectMeta{
				Name: "event1", Namespace: "demo-system",
			},
		}, time.Second*2)

		handler := SSE(ttlCache)
		req, err := http.NewRequest(http.MethodGet, "/notifications", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %v", rr.Code)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "text/event-stream" {
			t.Errorf("expected Content-Type 'text/event-stream', got %v", contentType)
		}

		got := string(rr.Body.Bytes())
		if !reflect.DeepEqual(got, exp) {
			t.Errorf("expected response body %v, got %v", exp, got)
		}
	})
}
