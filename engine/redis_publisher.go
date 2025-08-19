package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
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

	// subscriber fields
	group        string
	consumer     string
	ctx          context.Context
	cancel       context.CancelFunc
	onInvalidate func(string)
}

// newRedisPublisher creates a new Redis publisher with a stable instance name.
// instanceName should be a stable identifier (e.g., Pod name) to ensure consistent
// consumer group membership for tracking message consumption progress.
func newRedisPublisher(redisAddr, instanceName string) (*redisPublisher, error) {
	if redisAddr == "" {
		return nil, fmt.Errorf("redis address is required and cannot be empty")
	}
	if instanceName == "" {
		return nil, fmt.Errorf("instance name is required and cannot be empty")
	}
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithCancel(context.Background())
	p := &redisPublisher{
		client:   client,
		stream:   defaultStreamName,
		sourceID: instanceName, // Use instanceName instead of random ID
		msgs:     make(chan string, 1024),
		ctx:      ctx,
		cancel:   cancel,
	}
	p.wg.Add(1)
	go p.loop()
	return p, nil
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
	ctx := p.ctx
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

// startSubscriber creates a dedicated consumer group per instance and starts reading
// invalidation messages. Messages published by this instance are ignored.
func (p *redisPublisher) startSubscriber(onInvalidate func(string)) {
	if p == nil || p.client == nil {
		return
	}
	p.onInvalidate = onInvalidate
	// Use per-instance group and consumer so each instance receives all messages.
	p.group = fmt.Sprintf("sync-cache:grp:%s", p.sourceID)
	p.consumer = p.sourceID

	// Ensure group and consumer exist.
	if err := p.client.XGroupCreateMkStream(p.ctx, p.stream, p.group, "$").Err(); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "busygroup") {
			log.Printf("sync-cache: failed to create consumer group %s: %v", p.group, err)
		}
	}
	if err := p.client.XGroupCreateConsumer(p.ctx, p.stream, p.group, p.consumer).Err(); err != nil {
		// log only; consumer may already exist
		log.Printf("sync-cache: failed to create consumer %s: %v", p.consumer, err)
	}

	p.wg.Add(1)
	go p.subLoop()
}

func (p *redisPublisher) subLoop() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		streams, err := p.client.XReadGroup(p.ctx, &redis.XReadGroupArgs{
			Group:    p.group,
			Consumer: p.consumer,
			Streams:  []string{p.stream, ">"},
			Count:    128,
			Block:    5 * time.Second,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			// transient error; log and retry
			log.Printf("sync-cache: XReadGroup error: %v", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				// Parse fields
				rawKey, ok := msg.Values["key"]
				if !ok {
					// Ack and continue to avoid re-delivery storms
					_, _ = p.client.XAck(p.ctx, p.stream, p.group, msg.ID).Result()
					continue
				}
				key := fmt.Sprint(rawKey)
				if src, ok := msg.Values["src"]; ok {
					if fmt.Sprint(src) == p.sourceID {
						// Ignore self-published messages
						_, _ = p.client.XAck(p.ctx, p.stream, p.group, msg.ID).Result()
						continue
					}
				}

				// Invalidate locally
				if p.onInvalidate != nil {
					p.onInvalidate(key)
				}

				// Ack processed message
				if _, err := p.client.XAck(p.ctx, p.stream, p.group, msg.ID).Result(); err != nil {
					log.Printf("sync-cache: XAck failed for id=%s: %v", msg.ID, err)
				}
			}
		}
	}
}

func (p *redisPublisher) Close() error {
	if p == nil {
		return nil
	}
	if p.cancel != nil {
		p.cancel()
	}
	close(p.msgs)
	p.wg.Wait()
	return p.client.Close()
}
