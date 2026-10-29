package queue

import (
	"context"

	"github.com/akhand0ps/relay/internal/models"
)

type Queue interface {
	Enqueue(ctx context.Context, task models.DeliveryTask) error
	Dequeue(ctx context.Context) (*models.DeliveryTask, error)
	Ping(ctx context.Context) error
	Len(ctx context.Context) (int64, error)
	Close() error
}
