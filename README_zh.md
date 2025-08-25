# sync-cache

[English](README.md) | 中文

一个以最终一致性为目标的分布式 L1 本地缓存，通过 Redis Streams 异步广播失效消息来进行跨实例同步。

## 测试

### 环境搭建

```bash
cd tests/
docker compose up -d

# 验证服务健康状态
curl http://localhost:8080/health  # instance-a
curl http://localhost:8081/health  # instance-b

# 运行集成测试
go test ./integration/
```

### 手动测试示例

```bash
# 在 instance-a 中设置值
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"user:123","value":"John Doe"}'

# 从 instance-b 中获取值
curl "http://localhost:8081/get?key=user:123"

# 在 instance-a 中更新值
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"user:123","value":"Alice Doe"}'

# 从 instance-b 中获取同步后的值
curl "http://localhost:8081/get?key=user:123"
```

## FAQ

### 有哪些典型的缓存业务场景？分别适用于什么现有方案？

不同的业务场景适用不同的 Go 缓存方案。若需跨实例强一致性且能容忍网络开销，可直连 Redis。简单的单机应用可选 `go-cache`。对于需要缓存海量数据并规避 GC 停顿的单机场景，`big-cache` 是更优选择。而 `groupcache` 则专为分布式环境下的不可变数据缓存设计。

### 对于哪些业务场景，还没有很合适的解决方案？

现有方案难以满足一个普遍的微服务需求：当服务以多实例部署、需缓存可变数据，且同时要求极低的读取延迟和秒级的数据同步时，便出现了架构空白。单机缓存无法保证实例间数据一致，而为不可变数据设计的 `groupcache` 又不支持必要的数据失效机制，导致业务陷入两难。

### 这个项目尝试用什么方案解决这个问题？

本项目 `sync-cache` 提出一个两级（L1/L2）缓存方案。L1 是高性能的进程内缓存（如 Ristretto），保障极速读取性能。L2 是一个共享的失效总线（如 Redis Streams），它不存储业务数据，仅用于广播失效消息。当数据更新时，一个实例会通知总线，其他实例收到消息后便删除其本地 L1 缓存。

### 为什么实例需要稳定的身份标识？

每个缓存实例都需要一个稳定、唯一的标识符来维持在 Redis Streams 中的消费者组成员关系，以便跨重启追踪消息消费进度。没有稳定身份的实例重启后会创建新的消费者组，可能导致遗漏失效消息或丢失消费状态追踪。