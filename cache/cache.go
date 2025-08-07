package cache

type Key interface {
	uint64 | string | int | int32 | uint32 | int64
}

// InvalidationMessage defines the structure of invalidation messages broadcast via Redis Stream.
type InvalidationMessage[K Key] struct {
	Key       K      `json:"key"`
	Timestamp int64  `json:"ts"`  // UnixNano timestamp, for debugging and sorting
	SourceID  string `json:"src"` // Unique ID of the instance publishing the message
}

// SyncCache is a 2-tier, actively invalidated, thread-safe cache.
type SyncCache[K Key, V any] interface {
	// Get retrieves an item from the local L1 cache.
	Get(key K) (V, bool)

	// Set adds an item to the local L1 cache and broadcasts an invalidation message to other instances.
	// This operation is usually called after the data source is updated.
	Set(key K, value V)

	// Delete removes an item from the local L1 cache and broadcasts an invalidation message to other instances.
	Delete(key K)

	// Close gracefully closes the connection to the invalidation bus.
	Close() error
}
