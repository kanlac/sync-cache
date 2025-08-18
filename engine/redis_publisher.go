package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultStreamName = "sync-cache:invalidation"

type redisPublisher struct {
	client   *redis.Client
	stream   string
	sourceID string
	msgs     chan string
	wg       sync.WaitGroup
}

func newRedisPublisher(redisAddr string) *redisPublisher {
	if redisAddr == "" {
		panic("redis address is required and cannot be empty")
	}
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	p := &redisPublisher{
		client:   client,
		stream:   defaultStreamName,
		sourceID: generateSourceID(),
		msgs:     make(chan string, 1024),
	}
	p.wg.Add(1)
	go p.loop()
	return p
}

func generateSourceID() string {
	host, _ := os.Hostname()
	pid := os.Getpid()
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%d-%s", host, pid, hex.EncodeToString(b))
}

func (p *redisPublisher) publishInvalidation(key string) {
	if p == nil || p.client == nil {
		return
	}
	select {
	case p.msgs <- key:
	default:
		log.Printf("sync-cache: dropped invalidation publish due to full buffer, key=%s", key)
	}
}

func (p *redisPublisher) loop() {
	defer p.wg.Done()
	ctx := context.Background()
	for key := range p.msgs {
		if err := p.client.XAdd(ctx, &redis.XAddArgs{
			Stream: p.stream,
			Values: map[string]any{
				"key": key,
				"ts":  time.Now().UnixNano(),
				"src": p.sourceID,
			},
		}).Err(); err != nil {
			log.Printf("sync-cache: failed to publish invalidation, key=%s, err=%v", key, err)
		}
	}
}

func (p *redisPublisher) Close() error {
	if p == nil {
		return nil
	}
	close(p.msgs)
	p.wg.Wait()
	return p.client.Close()
}
