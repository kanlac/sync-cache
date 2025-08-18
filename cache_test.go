package main

import (
	"os"
	"testing"
)

func TestSyncCacheBasic(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set; skipping")
	}
	cache := NewSyncCache[string, string](addr)
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
