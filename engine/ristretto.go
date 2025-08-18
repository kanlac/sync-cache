package engine

import (
	"fmt"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/kanlac/sync-cache/cache"
)

type RistrettoEngine[K cache.Key, V any] struct {
	store     *ristretto.Cache[K, V]
	publisher *redisPublisher
}

func NewRistrettoCacheEngine[K cache.Key, V any](redisAddr string) *RistrettoEngine[K, V] {
	cache, err := ristretto.NewCache(&ristretto.Config[K, V]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	return &RistrettoEngine[K, V]{store: cache, publisher: newRedisPublisher(redisAddr)}
}

// Get retrieves a value from the Ristretto cache
func (r *RistrettoEngine[K, V]) Get(key K) (V, bool) {
	v, ok := r.store.Get(key)
	if !ok {
		var zero V
		return zero, false
	}
	return v, true
}

// Set sets a value in the Ristretto cache
func (r *RistrettoEngine[K, V]) Set(key K, value V) bool {
	// Ristretto requires a cost parameter, here simply set to 1
	ok := r.store.Set(key, value, 1)
	r.publisher.publishInvalidation(fmt.Sprint(key))
	return ok
}

// Delete removes a value from the Ristretto cache
func (r *RistrettoEngine[K, V]) Delete(key K) {
	r.store.Del(key)
	r.publisher.publishInvalidation(fmt.Sprint(key))
}

// Close shuts down the async publisher. Ristretto itself has no Close.
func (r *RistrettoEngine[K, V]) Close() error {
	if r.publisher != nil {
		return r.publisher.Close()
	}
	return nil
}
