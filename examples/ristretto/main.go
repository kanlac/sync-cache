package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kanlac/sync-cache/engine"
)

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		panic("REDIS_ADDR is required")
	}
	cache := engine.NewRistrettoCacheEngine[string, string](addr)
	defer cache.Close()

	ok := cache.Set("user:42", "Alice")
	if !ok {
		fmt.Println("Set returned false (possibly dropped)")
	}

	time.Sleep(10 * time.Millisecond)

	if v, found := cache.Get("user:42"); found {
		fmt.Println("Get user:42 =", v)
	} else {
		fmt.Println("user:42 not found")
	}

	cache.Delete("user:42")
	time.Sleep(10 * time.Millisecond)

	if _, found := cache.Get("user:42"); !found {
		fmt.Println("user:42 has been deleted")
	} else {
		fmt.Println("Delete failed: user:42 still exists")
	}
}
