package engine

import (
	"os"
	"testing"
	"time"
)

func TestRistrettoEngineBasic(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set; skipping")
	}
	e := NewRistrettoCacheEngine[string, string](addr)
	defer e.Close()

	ok := e.Set("foo", "bar")
	if !ok {
		t.Errorf("Set returned false, expected true")
	}
	time.Sleep(10 * time.Millisecond)
	v, found := e.Get("foo")
	if !found || v != "bar" {
		t.Errorf("expected Get to return 'bar', got '%v' (found=%v)", v, found)
	}

	e.Delete("foo")
	_, found = e.Get("foo")
	if found {
		t.Errorf("expected key to be deleted")
	}

	if err := e.Close(); err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}
