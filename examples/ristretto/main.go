package main

import (
	"fmt"
	"time"

	"github.com/kanlac/sync-cache/engine"
)

func main() {
	cache := engine.NewRistrettoCacheEngine[string, string]()
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
