package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/kanlac/sync-cache/cache"
	"github.com/redis/go-redis/v9"
)

type RistrettoEngine[K cache.Key, V any] struct {
	store       *ristretto.Cache[K, V]
	publisher   *redisPublisher
	redisClient *redis.Client
	keyPrefix   string
}

// NewRistrettoCacheEngine creates a new Ristretto-based cache engine with Redis invalidation.
// instanceName is required and should be a stable identifier for this instance (e.g., Pod name).
// This ensures consistent consumer group membership for tracking invalidation message consumption progress.
func NewRistrettoCacheEngine[K cache.Key, V any](redisAddr, instanceName string) (*RistrettoEngine[K, V], error) {
	if instanceName == "" {
		return nil, fmt.Errorf("instanceName is required and cannot be empty")
	}
	cache, err := ristretto.NewCache(&ristretto.Config[K, V]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}
	// Create Redis client for both L2 cache operations and publisher
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	
	publisher, err := newRedisPublisher(redisClient, instanceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis publisher: %w", err)
	}
	
	e := &RistrettoEngine[K, V]{
		store:       cache,
		publisher:   publisher,
		redisClient: redisClient,
		keyPrefix:   "sync-cache:data:",
	}
	e.publisher.startSubscriber(func(keyStr string) {
		if k, ok := parseKey[K](keyStr); ok {
			e.store.Del(k)
		}
	})
	return e, nil
}

// Get retrieves a value from the Ristretto L1 cache, falling back to Redis L2 cache on miss
func (r *RistrettoEngine[K, V]) Get(key K) (V, bool) {
	// Try L1 cache first
	v, ok := r.store.Get(key)
	if ok {
		return v, true
	}
	
	// L1 cache miss, try L2 Redis cache
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	redisKey := r.keyPrefix + fmt.Sprint(key)
	data, err := r.redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		// Redis miss or error
		var zero V
		return zero, false
	}
	
	// Deserialize from Redis
	var value V
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		// Deserialization error
		var zero V
		return zero, false
	}
	
	// Store in L1 cache for future hits
	r.store.Set(key, value, 1)
	
	return value, true
}

// Set sets a value in L1 Ristretto cache and asynchronously writes to L2 Redis cache
func (r *RistrettoEngine[K, V]) Set(key K, value V) {
	// Set in L1 cache immediately
	r.store.Set(key, value, 1)
	
	// Asynchronously set in L2 Redis cache
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		data, err := json.Marshal(value)
		if err == nil {
			redisKey := r.keyPrefix + fmt.Sprint(key)
			// Set with no expiration (0) - you can modify this if you want TTL
			r.redisClient.Set(ctx, redisKey, data, 0)
		}
	}()
	
	// Publish invalidation to other instances
	r.publisher.publishInvalidation(fmt.Sprint(key))
}

// Delete removes a value from L1 Ristretto cache and asynchronously deletes from L2 Redis cache
func (r *RistrettoEngine[K, V]) Delete(key K) {
	// Delete from L1 cache immediately
	r.store.Del(key)
	
	// Asynchronously delete from L2 Redis cache
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		redisKey := r.keyPrefix + fmt.Sprint(key)
		r.redisClient.Del(ctx, redisKey)
	}()
	
	// Publish invalidation to other instances
	r.publisher.publishInvalidation(fmt.Sprint(key))
}

// Close shuts down the async publisher and Redis client. Ristretto itself has no Close.
func (r *RistrettoEngine[K, V]) Close() error {
	var err error
	// Close publisher first (stops goroutines)
	if r.publisher != nil {
		err = r.publisher.Close()
	}
	// Then close the Redis client (managed by engine)
	if r.redisClient != nil {
		if closeErr := r.redisClient.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
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
