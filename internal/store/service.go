package store

import (
	"context"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"

	"github.com/go-redis/redis/v8"
)

type Service struct {
	cfg *model.Cfg
	log *logger.Logger
	KV  *KV
}

// KV redis storage object
type KV struct {
	Redis *redis.Client
}

// New creates a new instance of store
func New(ctx context.Context, cfg *model.Cfg, log *logger.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: log,
	}

	if err := s.newKV(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) newKV(ctx context.Context) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: s.cfg.Redis.Addr,
		DB:   s.cfg.Redis.DB,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}

	kv := &KV{
		Redis: redisClient,
	}

	s.KV = kv

	return nil
}
