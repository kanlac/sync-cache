package main

import "testing"

func TestSyncCacheBasic(t *testing.T) {
	cache := NewSyncCache[string, string]()
	defer cache.Close()

	cache.Set("foo", "bar")
	if v, ok := cache.Get("foo"); !ok || v != "bar" {
		t.Errorf("expected 'bar', got '%v' (ok=%v)", v, ok)
	}

	cache.Delete("foo")
	if _, ok := cache.Get("foo"); ok {
		t.Errorf("expected key to be deleted")
	}
}
