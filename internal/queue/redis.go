
package queue

import (
	"context"
	"encoding/json"
	"time"
	"github.com/redis/go-redis/v9"
	"github.com/akhand0ps/relay/internal/models"
)



type RedisQueue struct {

	client *redis.Client
	key string
}


func NewRedisQueue(addr string, key string) (*RedisQueue , error){

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})


	ctx,cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()


	if err:= client.Ping(ctx).Err(); err != nil{
		return nil,err
	}

	return &RedisQueue{
		client:client,
		key: key,
	},nil
}



func (rq * RedisQueue) Enqueue(ctx context.Context,task models.DeliveryTask) error {

	data,err := json.Marshal(task)



	if err != nil{
		return err
	}
	

	return rq.client.LPush(ctx,rq.key,data).Err()

}