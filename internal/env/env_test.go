package env

import (
	"os"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	v1 := String("K1", "def")
	if v1 != "def" {
		t.Fatal("value 'v1' should be 'def'")
	}

	if err := os.Setenv("K1", "abc"); err != nil {
		t.Fatal(err)
	}

	v1 = String("K1", "def")
	if v1 != "abc" {
		t.Fatal("value 'v1' should be 'abc'")
	}
}

func TestInt(t *testing.T) {
	v1 := Int("K1", 1)
	if v1 != 1 {
		t.Fatal("value 'v1' should be '1'")
	}

	if err := os.Setenv("K1", "8"); err != nil {
		t.Fatal(err)
	}

	v1 = Int("K1", 1)
	if v1 != 8 {
		t.Fatal("value 'v1' should be '8'")
	}
}

func TestBool(t *testing.T) {
	v1 := Bool("K1", false)
	if v1 != false {
		t.Fatal("value 'v1' should be 'false'")
	}

	if err := os.Setenv("K1", "true"); err != nil {
		t.Fatal(err)
	}

	v1 = Bool("K1", false)
	if v1 != true {
		t.Fatal("value 'v1' should be 'true'")
	}
}

func TestDuration(t *testing.T) {
	v1 := Duration("K1", time.Second*2)
	if v1.String() != "2s" {
		t.Fatal("value 'v1' should be '2s'")
	}

	if err := os.Setenv("K1", "4m"); err != nil {
		t.Fatal(err)
	}

	v1 = Duration("K1", time.Second*2)
	if v1.String() != "4m0s" {
		t.Fatal("value 'v1' should be '4m0s'")
	}
}
