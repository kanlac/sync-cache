# sync-cache

English | [中文](README_zh.md)

An in-memory L1 cache for Go that targets eventual consistency across instances via asynchronous invalidation with Redis Streams.

## FAQ

### What are typical caching scenarios and their existing solutions?

Different Go caching solutions fit different scenarios. For strong cross-instance consistency where network latency is acceptable, a direct Redis connection can be used. For simple, single-machine applications, `go-cache` is suitable. `big-cache` is a better choice for single-node contexts that cache massive datasets and need to avoid GC pauses. `groupcache` is specifically designed for caching immutable data in a distributed environment.

### For which business scenarios is there no ideal solution?

A common architectural gap exists for modern microservices. When a service is deployed across multiple instances and needs to cache mutable data, it faces a conflict: it requires both the low-latency reads of an in-process cache and near real-time data consistency. Single-node caches cannot ensure consistency across instances, while `groupcache`'s design for immutable data lacks the required invalidation mechanism.

### How does this project attempt to solve this problem?

This project, `sync-cache`, proposes a two-tier (L1/L2) architecture. The L1 cache is a high-performance, in-process store (like Ristretto) in each service instance, ensuring ultra-fast reads. The L2 is a shared invalidation bus (like Redis Streams) that does not store business data; its sole purpose is to broadcast invalidation messages. When data is updated on one instance, it notifies the bus, and all other instances receive a message to evict the item from their local L1 cache. 