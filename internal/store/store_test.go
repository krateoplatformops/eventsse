package store

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/krateoplatformops/eventsse/internal/cache"
	"github.com/krateoplatformops/eventsse/internal/labels"

	corev1 "k8s.io/api/core/v1"
)

func TestClientTTL(t *testing.T) {
	var c TTLSetter = &Client{}
	c.SetTTL(200)
}

func TestClientPrepareKey(t *testing.T) {
	const exp = "events/comp-abc/123"

	var c KeyPreparer = &Client{}
	got := c.PrepareKey("123", "abc")
	if got != exp {
		t.Fatalf("ttl: got %v, expected %v", got, exp)
	}
}

func TestGet(t *testing.T) {
	var sto Store
	if len(os.Getenv("INTEGRATION")) > 0 {
		//t.Skip("skipping integration tests: set INTEGRATION environment variable")
		var err error
		sto, err = NewClient(DefaultOptions)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		sto = &MockStore{
			ttl:  time.Second * 10,
			data: cache.NewTTL[string, corev1.Event](),
		}
	}
	defer sto.Close()

	key := sto.PrepareKey("", "abcde12345")
	_, ok, err := sto.Get(key, GetOptions{
		Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected no data")
	}
}

func TestPut(t *testing.T) {
	var sto Store
	if len(os.Getenv("INTEGRATION")) > 0 {
		//t.Skip("skipping integration tests: set INTEGRATION environment variable")
		var err error
		sto, err = NewClient(DefaultOptions)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		sto = &MockStore{
			ttl:  time.Second * 10,
			data: cache.NewTTL[string, corev1.Event](),
		}
	}
	defer sto.Close()

	files := []string{
		"../../testdata/event.sample1.json",
		"../../testdata/event.sample2.json",
	}

	for _, x := range files {
		fin, err := os.Open(x)
		if err != nil {
			t.Fatal(err)
		}
		defer fin.Close()

		var nfo corev1.Event
		if err := json.NewDecoder(fin).Decode(&nfo); err != nil {
			t.Fatal(err)
		}

		key := sto.PrepareKey(string(nfo.UID), labels.CompositionID(&nfo))
		t.Logf("key: %s", key)

		err = sto.Set(key, &nfo)
		if err != nil {
			t.Fatal(err)
		}
	}
}

var _ Store = (*MockStore)(nil)

// MockStore è un mock del client store per testare l'handler
type MockStore struct {
	data *cache.TTLCache[string, corev1.Event]
	ttl  time.Duration
}

func (m *MockStore) PrepareKey(uid, compositionID string) string {
	return uid + ":" + compositionID
}

func (m *MockStore) Set(key string, event *corev1.Event) error {
	m.data.Set(key, *event, m.ttl)
	return nil
}

func (m *MockStore) Get(key string, opts GetOptions) (data []corev1.Event, found bool, err error) {
	obj, exists := m.data.Get(key)
	if !exists {
		return nil, false, nil
	}
	return []corev1.Event{obj}, true, nil
}

func (m *MockStore) Delete(key string) error {
	m.data.Pop(key)
	return nil
}

func (m *MockStore) SetTTL(x int) {
	m.ttl = time.Second * time.Duration(x)
}

func (m *MockStore) Close() error {
	m.data.Clear()
	return nil
}
