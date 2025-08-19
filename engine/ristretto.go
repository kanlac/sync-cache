package engine

import (
	"fmt"
	"strconv"

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
	e := &RistrettoEngine[K, V]{store: cache, publisher: newRedisPublisher(redisAddr)}
	e.publisher.startSubscriber(func(keyStr string) {
		if k, ok := parseKey[K](keyStr); ok {
			e.store.Del(k)
		}
	})
	return e
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

func parseKey[K cache.Key](s string) (K, bool) {
	var zero K
	switch any(zero).(type) {
	case string:
		return any(s).(K), true
	case int:
		v, err := strconv.Atoi(s)
		if err != nil {
			return zero, false
		}
		return any(v).(K), true
	case int32:
		v64, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return zero, false
		}
		return any(int32(v64)).(K), true
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return zero, false
		}
		return any(v).(K), true
	case uint32:
		v64, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return zero, false
		}
		return any(uint32(v64)).(K), true
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return zero, false
		}
		return any(v).(K), true
	default:
		return zero, false
	}
}
