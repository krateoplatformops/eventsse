package store

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/eventsse/internal/labels"

	corev1 "k8s.io/api/core/v1"
)

func TestClientTTL(t *testing.T) {
	var c TTLSetter = &Client{}
	c.SetTTL(200)
}

func TestClientPrepareKey(t *testing.T) {
	const exp = "krateo.io.events/comp-abc/123"

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
		sto = &MockStore{}
	}
	defer sto.Close()

	key := sto.PrepareKey("", "abcde12345")
	_, ok, _ := sto.Get(key, GetOptions{
		Limit: 10,
	})
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
		sto = &MockStore{}
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

func TestKeys(t *testing.T) {
	var sto Store
	if len(os.Getenv("INTEGRATION")) > 0 {
		//t.Skip("skipping integration tests: set INTEGRATION environment variable")
		var err error
		sto, err = NewClient(DefaultOptions)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		sto = &MockStore{}
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

	keys, err := sto.Keys(0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s\n", keys)
}

var _ Store = (*MockStore)(nil)

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

func (m *MockStore) Get(key string, opts GetOptions) (data []corev1.Event, found bool, err error) {
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

func (m *MockStore) SetTTL(_ int) {}

func (m *MockStore) Close() error { return nil }

func (m *MockStore) Keys(l int) ([]string, error) {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}
