package engine

// SimpleEngine is a simple implementation based on map
type SimpleEngine[K comparable, V any] struct {
	store map[K]V
}

func NewSimpleEngine[K comparable, V any]() *SimpleEngine[K, V] {
	return &SimpleEngine[K, V]{store: make(map[K]V)}
}

func (m *SimpleEngine[K, V]) Get(key K) (V, bool) {
	v, ok := m.store[key]
	return v, ok
}

func (m *SimpleEngine[K, V]) Set(key K, value V) bool {
	m.store[key] = value
	return true
}

func (m *SimpleEngine[K, V]) Delete(key K) {
	delete(m.store, key)
}

func (m *SimpleEngine[K, V]) Close() error {
	return nil
}
