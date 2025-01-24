package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"solution/internal/adapters/logger"
	"time"
)

type tokenRedisStorage struct {
	db *redis.Client
}

func NewTokenStorage(db *redis.Client) *tokenRedisStorage {
	return &tokenRedisStorage{db: db}
}

func (s *tokenRedisStorage) SetAuthID(ctx context.Context, userID, authID string) error {
	return s.db.Set(ctx, userID, authID, time.Hour).Err()
}

func (s *tokenRedisStorage) CheckAuthID(ctx context.Context, userID, authID string) (bool, error) {
	authIDFromRedis, err := s.db.Get(ctx, userID).Result()
	if err != nil {
		logger.Log.Error(err.Error())
		return false, err
	}

	return authIDFromRedis == authID, nil
}
