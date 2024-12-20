package redis

import (
	"context"
	"fmt"
	"task/config"
	"task/storage"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	db *redis.Client
}

func New(cfg config.Config) storage.IRedisStorage {
	var redisClient *redis.Client

	// Check if Redis URL is provided
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			panic(fmt.Sprintf("Invalid Redis URL: %v", err))
		}
		redisClient = redis.NewClient(opt)
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr: cfg.RedisHost + ":" + cfg.RedisPort,
		})
	}

	return Store{
		db: redisClient,
	}
}

func (s Store) SetX(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	StatusCmd := s.db.SetEx(ctx, key, value, duration)
	if StatusCmd.Err() != nil {
		return StatusCmd.Err()
	}

	return nil
}

func (s Store) Get(ctx context.Context, key string) (interface{}, error) {
	resp := s.db.Get(ctx, key)

	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return resp.Val(), nil
}

func (s Store) Del(ctx context.Context, key string) error {
	StatusCmd := s.db.Del(ctx, key)
	if StatusCmd.Err() != nil {
		return StatusCmd.Err()
	}
	return nil
}
