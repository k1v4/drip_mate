package redispkg

import (
	"context"
	"fmt"

	"github.com/k1v4/drip_mate/internal/config"
	"github.com/k1v4/drip_mate/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, cfg *config.RedisConfig) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	l := logger.GetLoggerFromContext(ctx)

	pong, err := db.Ping(ctx).Result()
	if err != nil {
		l.Error(ctx, fmt.Sprintf("failed to connect to redis server: %s\n", err.Error()))
		return nil, err
	}

	l.Info(ctx, fmt.Sprintf("%s connected to redis server: %s\n", pong, cfg.Port))

	return db, nil
}
