package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type activationRedisStorage struct {
	db *redis.Client
}

func NewActivationStorage(db *redis.Client) *activationRedisStorage {
	return &activationRedisStorage{db: db}
}

func (s *activationRedisStorage) Cache(ctx context.Context, email string, until time.Time) error {
	duration, _ := time.ParseDuration("24h")
	return s.db.Set(ctx, email, until, duration).Err()
}

func (s *activationRedisStorage) CheckCache(ctx context.Context, email string) (bool, error) {
	until, err := s.db.Get(ctx, email).Time()
	if err != nil {
		return false, err
	}
	if time.Now().After(until) {
		return false, nil
	}
	return true, nil
}
