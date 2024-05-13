package main

import (
	"context"
	"ip_service/internal/apiv1"
	"ip_service/internal/httpserver"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/pkg/configuration"
	"ip_service/pkg/logger"
	"ip_service/pkg/trace"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type service interface {
	Close(ctx context.Context) error
}

func main() {
	var wg sync.WaitGroup
	ctx := context.Background()

	services := make(map[string]service)

	cfg, err := configuration.Parse(ctx, logger.NewSimple("Configuration"))
	if err != nil {
		panic(err)
	}

	log, err := logger.New("ip_service", cfg.IPService.Log.FolderPath, cfg.IPService.Production)
	if err != nil {
		panic(err)
	}

	tracer, err := trace.New(ctx, cfg, log, "ip_service", "kalle_kula")
	if err != nil {
		panic(err)
	}

	store, err := store.New(ctx, cfg, tracer, log.New("store"))
	services["store"] = store
	if err != nil {
		panic(err)
	}
	max, err := maxmind.New(ctx, cfg, store, tracer, log.New("maxmind"))
	services["maxmind"] = max
	if err != nil {
		panic(err)
	}
	apiv1, err := apiv1.New(ctx, max, store, cfg, tracer, log.New("apiv1"))
	if err != nil {
		panic(err)
	}
	httpserver, err := httpserver.New(ctx, cfg, apiv1, tracer, log.New("httpserver"))
	services["httpserver"] = httpserver
	if err != nil {
		panic(err)
	}

	// add timestamp in storage which time the service was started
	if err := store.KV.Set(ctx, "started", time.Now().String()); err != nil {
		panic(err)
	}

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan // Blocks here until interrupted
	mainLog := log.New("main")
	mainLog.Info("HALTING SIGNAL!")

	for serviceName, service := range services {
		if err := service.Close(ctx); err != nil {
			mainLog.Error(err, "serviceName", serviceName)
		}
	}

	wg.Wait() // Block here until are workers are done

	mainLog.Info("Stopped")

}
