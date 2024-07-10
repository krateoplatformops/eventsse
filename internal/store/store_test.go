package store

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/eventsse/internal/labels"
	v1 "k8s.io/api/core/v1"
)

func TestGet(t *testing.T) {
	sto, err := NewClient(DefaultOptions)
	if err != nil {
		t.Fatal(err)
	}
	defer sto.Close()

	key := sto.PrepareKey("", "abcde12345")
	all, _, err := sto.Get(key, GetOptions{
		Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	dat, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(dat))
}

func TestPut(t *testing.T) {
	sto, err := NewClient(DefaultOptions)
	if err != nil {
		t.Fatal(err)
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

		var nfo v1.Event
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
