package main

import (
	"github.com/kanlac/sync-cache/cache"
	"github.com/kanlac/sync-cache/engine"
)

type syncCacheImpl[K cache.Key, V any] struct {
	e *engine.SimpleEngine[K, V]
}

// NewSyncCache creates a new sync cache with the given Redis address and instance name.
// instanceName should be a stable identifier for this instance (e.g., Pod name).
func NewSyncCache[K cache.Key, V any](redisAddr, instanceName string) (cache.SyncCache[K, V], error) {
	e, err := engine.NewSimpleEngine[K, V](redisAddr, instanceName)
	if err != nil {
		return nil, err
	}
	return &syncCacheImpl[K, V]{e: e}, nil
}

func (s *syncCacheImpl[K, V]) Get(key K) (V, bool) {
	return s.e.Get(key)
}

func (s *syncCacheImpl[K, V]) Set(key K, value V) {
	s.e.Set(key, value)
}

func (s *syncCacheImpl[K, V]) Delete(key K) {
	s.e.Delete(key)
}

func (s *syncCacheImpl[K, V]) Close() error {
	return s.e.Close()
}
