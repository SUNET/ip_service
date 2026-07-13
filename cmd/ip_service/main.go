package main

import (
	"context"
	"ip_service/internal/apiv1"
	"ip_service/internal/httpserver"
	"ip_service/internal/lctree"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/internal/whois"
	"ip_service/pkg/configuration"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/model"
	"github.com/SUNET/vc/pkg/trace"
)

type service interface {
	Close(ctx context.Context) error
}

func main() {
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

	tracer, err := trace.New(ctx, &model.Cfg{
		Common: &model.Common{
			Tracing: model.OTEL{
				Enable: cfg.IPService.Tracing.Enable,
				Addr:   cfg.IPService.Tracing.Addr,
			},
		},
	}, "ip_service", log)
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

	ipTree := lctree.New(log.New("lctree"))
	services["lctree"] = ipTree

	whoisService, err := whois.New(ctx, cfg, ipTree, store, log.New("whois"))
	services["whois"] = whoisService
	if err != nil {
		panic(err)
	}

	// Force GC and return memory to OS after loading large RPSL datasets
	runtime.GC()
	debug.FreeOSMemory()

	apiv1, err := apiv1.New(ctx, max, whoisService, store, cfg, tracer, log.New("apiv1"))
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

	mainLog := log.New("main")
	mainLog.Info("All services started")

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	mainLog.Info("HALTING SIGNAL!")

	for serviceName, service := range services {
		if err := service.Close(ctx); err != nil {
			mainLog.Error(err, "serviceName", serviceName)
		}
	}

	stop()

	mainLog.Info("Stopped")

}
