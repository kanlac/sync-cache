package engine

import (
	"fmt"
	"strconv"
)

// SimpleEngine is a simple implementation based on map
type SimpleEngine[K comparable, V any] struct {
	store     map[K]V
	publisher *redisPublisher
}

// NewSimpleEngine creates a new simple map-based cache engine with Redis invalidation.
// instanceName is required and should be a stable identifier for this instance (e.g., Pod name).
// This ensures consistent consumer group membership for tracking invalidation message consumption progress.
func NewSimpleEngine[K comparable, V any](redisAddr, instanceName string) (*SimpleEngine[K, V], error) {
	if instanceName == "" {
		return nil, fmt.Errorf("instanceName is required and cannot be empty")
	}
	publisher, err := newRedisPublisher(redisAddr, instanceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis publisher: %w", err)
	}
	e := &SimpleEngine[K, V]{store: make(map[K]V), publisher: publisher}
	e.publisher.startSubscriber(func(keyStr string) {
		if k, ok := parseKeySimple[K](keyStr); ok {
			delete(e.store, k)
		}
	})
	return e, nil
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

func parseKeySimple[K comparable](s string) (K, bool) {
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
