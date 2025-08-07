package main

import (
	"github.com/kanlac/sync-cache/cache"
	"github.com/kanlac/sync-cache/engine"
)

type syncCacheImpl[K cache.Key, V any] struct {
	e *engine.SimpleEngine[K, V]
}

func NewSyncCache[K cache.Key, V any]() cache.SyncCache[K, V] {
	return &syncCacheImpl[K, V]{e: engine.NewSimpleEngine[K, V]()}
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
