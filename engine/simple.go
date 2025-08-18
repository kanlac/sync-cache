package engine

import (
	"fmt"
)

// SimpleEngine is a simple implementation based on map
type SimpleEngine[K comparable, V any] struct {
	store     map[K]V
	publisher *redisPublisher
}

func NewSimpleEngine[K comparable, V any](redisAddr string) *SimpleEngine[K, V] {
	return &SimpleEngine[K, V]{store: make(map[K]V), publisher: newRedisPublisher(redisAddr)}
}

func (m *SimpleEngine[K, V]) Get(key K) (V, bool) {
	v, ok := m.store[key]
	return v, ok
}

func (m *SimpleEngine[K, V]) Set(key K, value V) bool {
	m.store[key] = value
	m.publisher.publishInvalidation(fmt.Sprint(key))
	return true
}

func (m *SimpleEngine[K, V]) Delete(key K) {
	delete(m.store, key)
	m.publisher.publishInvalidation(fmt.Sprint(key))
}

func (m *SimpleEngine[K, V]) Close() error {
	if m.publisher != nil {
		return m.publisher.Close()
	}
	return nil
}
