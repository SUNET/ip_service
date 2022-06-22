package main

import (
	"context"
	"ip_service/internal/apiv1"
	"ip_service/internal/httpserver"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/pkg/configuration"
	"ip_service/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type service interface {
	Close(ctx context.Context) error
}

func main() {
	ctx := context.Background()
	wg := &sync.WaitGroup{}

	var (
		log      *logger.Logger
		mainLog  *logger.Logger
		services = make(map[string]service)
	)

	cfg, err := configuration.Parse(logger.NewSimple("Configuration"))
	if err != nil {
		panic(err)
	}

	mainLog = logger.New("main", cfg.Production)
	log = logger.New("ip_service", cfg.Production)

	store, err := store.New(ctx, cfg, log.New("store"))
	if err != nil {
		panic(err)
	}
	max, err := maxmind.New(ctx, cfg, store, log.New("maxmind"))
	if err != nil {
		panic(err)
	}
	apiv1, err := apiv1.New(ctx, max, cfg, log.New("apiv1"))
	if err != nil {
		panic(err)
	}
	httpserver, err := httpserver.New(ctx, cfg, apiv1, log.New("httpserver"))
	services["httpserver"] = httpserver
	if err != nil {
		panic(err)
	}

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan // Blocks here until interrupted

	log.Info("HALTING SIGNAL!")

	for serviceName, service := range services {
		if err := service.Close(ctx); err != nil {
			mainLog.Warn("shutdown", serviceName, err)
		}
	}

	wg.Wait() // Block here until are workers are done

	mainLog.Info("Stopped")

}
