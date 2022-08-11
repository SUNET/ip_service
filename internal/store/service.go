package store

import (
	"context"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"time"

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

	kv := &KV{
		File: diskvClient,
	}

	s.KV = kv

	return nil
}

var (
	nextRun    time.Time
	lastStatus *model.StatusService
)

func (s *Service) Status(ctx context.Context) *model.StatusService {
	if time.Now().After(nextRun) {
		status := &model.StatusService{
			ServiceName: "store",
			Timestamp:   time.Now(),
			Interval:    10 * time.Second,
		}

		if err := s.KV.Set(ctx, "statusTest", "test"); err != nil {
			status.Message = err.Error()
			return status
		}

		if s.KV.Get(ctx, "statusTest") != "test" {
			status.Message = "Can't access statusTest key-value entry"
			return status
		}

		if err := s.KV.Del(ctx, "statusTest"); err != nil {
			status.Message = err.Error()
			return status
		}

		status.Healthy = true

		nextRun = time.Now().Add(status.Interval)
		lastStatus = status

		return status
	}

	return lastStatus
}

// Close closes store service
func (s *Service) Close(ctx context.Context) error {
	s.log.Info("Quit")
	ctx.Done()
	return nil
}
