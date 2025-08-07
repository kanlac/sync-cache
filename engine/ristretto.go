package engine

import (
	"github.com/dgraph-io/ristretto/v2"
	"github.com/kanlac/sync-cache/cache"
)

type RistrettoEngine[K cache.Key, V any] struct {
	store *ristretto.Cache[K, V]
}

func NewRistrettoCacheEngine[K cache.Key, V any]() *RistrettoEngine[K, V] {
	cache, err := ristretto.NewCache(&ristretto.Config[K, V]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	return &RistrettoEngine[K, V]{store: cache}
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
	return r.store.Set(key, value, 1)
}

// Delete removes a value from the Ristretto cache
func (r *RistrettoEngine[K, V]) Delete(key K) {
	r.store.Del(key)
}

// Close closes the Ristretto cache (no-op)
func (r *RistrettoEngine[K, V]) Close() error {
	// Ristretto does not have an explicit Close method, just return nil
	return nil
}
