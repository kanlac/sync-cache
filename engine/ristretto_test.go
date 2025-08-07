package engine

import (
	"testing"
	"time"
)

func TestRistrettoEngineBasic(t *testing.T) {
	e := NewRistrettoCacheEngine[string, string]()
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
