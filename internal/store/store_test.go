package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/krateoplatformops/eventsse/internal/labels"
	v1 "k8s.io/api/core/v1"
)

func TestPut(t *testing.T) {
	fin, err := os.Open("../../testdata/event.sample.json")
	if err != nil {
		t.Fatal(err)
	}
	defer fin.Close()

	// err = yaml.NewYAMLOrJSONDecoder(fin, 100).Decode(&nfo)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	var nfo v1.Event
	if err := json.NewDecoder(fin).Decode(&nfo); err != nil {
		t.Fatal(err)
	}

	sto, err := NewClient(DefaultOptions)
	if err != nil {
		t.Fatal(err)
	}
	defer sto.Close()

	sto.SetTTL(2)

	key := nfo.FirstTimestamp.Format("20060102150405.00")
	if val, ok := labels.CompositionID(&nfo); ok {
		key = path.Join(key, val)
	}
	key = path.Join(key, string(nfo.UID))
	t.Logf("key: %s", key)

	err = sto.Set(key, &nfo)
	if err != nil {
		t.Fatal(err)
	}

	all, _, err := sto.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	dat, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(dat))
}
