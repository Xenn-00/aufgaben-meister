package worker

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func asynqRedisOpt(redis *redis.Client) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     redis.Options().Addr,
		Password: redis.Options().Password,
		DB:       redis.Options().DB,
	}
}
