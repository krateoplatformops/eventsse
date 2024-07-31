package cache

import (
	"testing"
	"time"
)

func TestTTLCache(t *testing.T) {
	c := NewTTL[string, int]()

	c.Set("one", 1, 1*time.Second)
	c.Set("two", 2, 2*time.Second)
	c.Set("three", 3, 4*time.Second)

	if _, found := c.Get("two"); !found {
		t.Fatal("Key 'two' not found in the cache or has expired")
	}

	if l := len(c.Keys()); l != 3 {
		t.Fatalf("Found: %d keys, expected: 3", l)
	}

	// Wait for a while to allow some items to expire
	time.Sleep(3 * time.Second)

	// Try to retrieve an expired key
	if _, found := c.Get("one"); found {
		t.Fatal("key 'one': should be expired")
	}

	// Pop a key from the cache
	if _, found := c.Pop("two"); found {
		t.Fatal("key 'two': should be expired")
	}

	if _, found := c.Pop("three"); !found {
		t.Fatal("key 'three': should NOT be expired")
	}

	c.Remove("three")

	c.Clear()
}
