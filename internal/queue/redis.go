package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/akhand0ps/relay/internal/models"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(addr string, key string) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisQueue{
		client: client,
		key:    key,
	}, nil
}

func (rq *RedisQueue) Enqueue(ctx context.Context, task models.DeliveryTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return rq.client.LPush(ctx, rq.key, data).Err()
}

func (rq *RedisQueue) Dequeue(ctx context.Context) (*models.DeliveryTask, error) {
	// BRPop blocks until an item is available or context is cancelled
	res, err := rq.client.BRPop(ctx, 0, rq.key).Result()
	if err != nil {
		return nil, err
	}

	// res[0] is key name, res[1] is the JSON payload
	var task models.DeliveryTask
	if err := json.Unmarshal([]byte(res[1]), &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (rq *RedisQueue) Ping(ctx context.Context) error {
	return rq.client.Ping(ctx).Err()
}

func (rq *RedisQueue) Len(ctx context.Context) (int64, error) {
	return rq.client.LLen(ctx, rq.key).Result()
}

func (rq *RedisQueue) Close() error {
	return rq.client.Close()
}
