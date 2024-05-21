package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/bullgare/pow-ddos-protection/internal/domain/contracts"
)

const (
	redisTTL      = 2 * time.Minute
	dataSeparator = "|"
)

func NewAuthStorage(redisAddress string) (*AuthStorage, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	// we ping here as it's one of the crucial parts of the app.
	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", redisAddress, err)
	}

	return &AuthStorage{
		redisClient: redisClient,
		ttl:         redisTTL,
	}, nil
}

var _ contracts.AuthStorage = &AuthStorage{}

type AuthStorage struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func (s *AuthStorage) Store(ctx context.Context, data contracts.AuthData) error {
	cmd := s.redisClient.Set(ctx, s.generateKey(data), s.generateValue(data), s.ttl)
	return cmd.Err()
}

func (s *AuthStorage) Delete(ctx context.Context, data contracts.AuthData) error {
	cmd := s.redisClient.Del(ctx, s.generateKey(data))
	return cmd.Err()
}

func (s *AuthStorage) CheckExists(ctx context.Context, data contracts.AuthData) (bool, error) {
	value, err := s.redisClient.Get(ctx, s.generateKey(data)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return value == s.generateValue(data), nil
}

func (s *AuthStorage) generateKey(data contracts.AuthData) string {
	return data.UserID
}

func (s *AuthStorage) generateValue(data contracts.AuthData) string {
	return data.UserID + dataSeparator + data.Seed
}
