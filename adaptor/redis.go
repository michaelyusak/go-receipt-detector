package adaptor

import (
	"github.com/michaelyusak/go-helper/adaptor"
	"github.com/michaelyusak/go-helper/entity"
	"github.com/redis/go-redis/v9"
)

func ConnectRedis(config entity.RedisConfig) *redis.Client {
	return adaptor.ConnectRedis(config)
}
