package store

import (
	"context"
	"fmt"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"

	"github.com/peterbourgon/diskv/v3"
)

type Service struct {
	cfg *model.Cfg
	log *logger.Logger
	KV  *KV
}

// KV key-value storage object
type KV struct {
	File *diskv.Diskv
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
	diskvClient := diskv.New(diskv.Options{
		BasePath:     "store",
		Transform:    func(s string) []string { return []string{} },
		CacheSizeMax: 1024 * 1024,
	})
	fmt.Println("client", diskvClient.WriteString("test", "test"))

	kv := &KV{
		File: diskvClient,
	}

	s.KV = kv

	return nil
}
