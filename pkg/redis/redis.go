package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Config struct {
	EnableCache bool   `toml:"enable_cache"`
	Address     string `toml:"address"`
	Password    string `toml:"password"`
}

type redis struct {
	Client *goredis.Client
}

type Client interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
}

func (redis *redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	return redis.Client.Set(ctx, key, value, expiration)
}

func (redis *redis) Get(ctx context.Context, key string) *goredis.StringCmd {
	return redis.Client.Get(ctx, key)
}

type redisClientMock struct{}

func (mock *redisClientMock) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	return nil
}

func (mock *redisClientMock) Get(ctx context.Context, key string) *goredis.StringCmd {
	return nil
}

func NewClient(redisCfg Config) (Client, error) {
	if redisCfg.EnableCache {
		rdb := goredis.NewClient(&goredis.Options{
			Addr:     redisCfg.Address,
			Password: redisCfg.Password,
			DB:       0, // use default DB
		})

		return rdb, nil
	} else {
		return &redisClientMock{}, nil
	}
}
