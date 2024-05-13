package store

import (
	"context"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/trace"
	"strings"
	"time"

	"github.com/peterbourgon/diskv/v3"
)

// Service store service
type Service struct {
	cfg        *model.Cfg
	log        *logger.Log
	TP         *trace.Tracer
	KV         *KV
	probeStore *model.StatusProbeStore
}

// KV key-value storage object
type KV struct {
	File *diskv.Diskv
}

// New creates a new instance of store
func New(ctx context.Context, cfg *model.Cfg, tp *trace.Tracer, log *logger.Log) (*Service, error) {
	s := &Service{
		cfg:        cfg,
		log:        log,
		TP:         tp,
		probeStore: &model.StatusProbeStore{},
	}

	if err := s.newKV(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func advancedTransform(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last] + ".txt",
	}
}

func inverseTransform(pathKey *diskv.PathKey) (key string) {
	txt := pathKey.FileName[len(pathKey.FileName)-4:]
	if txt != ".txt" {
		panic("Invalid file found in storage folder!")
	}
	return strings.Join(pathKey.Path, "/") + pathKey.FileName[:len(pathKey.FileName)-4]
}

func (s *Service) newKV(ctx context.Context) error {
	diskvClient := diskv.New(diskv.Options{
		BasePath:          s.cfg.IPService.Store.File.Path,
		AdvancedTransform: advancedTransform,
		InverseTransform:  inverseTransform,
		CacheSizeMax:      1024 * 1024,
	})

	kv := &KV{
		File: diskvClient,
	}

	s.KV = kv

	return nil
}

// Status returns the status of the service
func (s *Service) Status(ctx context.Context) *model.StatusProbe {
	if time.Now().Before(s.probeStore.NextCheck) {
		return s.probeStore.PreviousResult
	}

	probe := &model.StatusProbe{
		Name:          "kv",
		Healthy:       true,
		Message:       map[string]any{"status": "ok"},
		LastCheckedTS: time.Now(),
	}

	if err := s.KV.statusTest(ctx); err != nil {
		probe.Healthy = false
		probe.Message["status"] = err.Error()

	}

	s.probeStore.PreviousResult = probe
	s.probeStore.NextCheck = time.Now().Add(10 * time.Second)

	return probe
}

// Close closes store service
func (s *Service) Close(ctx context.Context) error {
	s.log.Info("Quit")
	ctx.Done()
	return nil
}
