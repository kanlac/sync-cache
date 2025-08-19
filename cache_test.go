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
	cache, err := NewSyncCache[string, string](addr, "test-cache-instance")
	if err != nil {
		t.Fatalf("Failed to create sync cache: %v", err)
	}
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
