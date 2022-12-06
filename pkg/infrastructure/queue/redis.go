package queue

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
)

type RedisQueue struct {
	store *redis.Client
	log   logger.Logger
}

func (r *RedisQueue) Push(ctx context.Context, queue string, item string) {
	result, err := json.MarshalNoEscape(&item)

	if err != nil {
		r.log.Errorf("Cannot push item to queue! %v\n", err)
		return
	}

	r.store.LPush(ctx, queue, result)
}

func (r *RedisQueue) Pop(ctx context.Context, queue string) string {
	return r.store.BRPop(ctx, 0, queue).Val()[1]
}

func (r *RedisQueue) Subscribe(ctx context.Context, queue string) <-chan string {
	channel := make(chan string)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(channel)
				return
			default:
				channel <- r.Pop(ctx, queue)
			}
		}
	}()
	return channel
}

func (r *RedisQueue) Len(ctx context.Context, queue string) int {
	return int(r.store.LLen(ctx, queue).Val())
}

func New(store *redis.Client, log logger.Logger) *RedisQueue {
	return &RedisQueue{store, log}
}
