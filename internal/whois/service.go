package whois

import (
	"context"
	"ip_service/internal/lctree"
	"ip_service/internal/rpslsource"
	"ip_service/internal/store"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/rpsl"
	"sync"
	"time"
)

// Service struct handles whois service
type Service struct {
	cfg             *model.Cfg
	log             *logger.Log
	wg              sync.WaitGroup
	updateTicker    time.Ticker
	ripe            *rpslsource.Service
	radb            *rpslsource.Service
	store           *store.Service
	RPSLRouterClass rpsl.RouterClass
	mu              sync.RWMutex
	tree            *lctree.Service
}

// New creates a new whois service
func New(ctx context.Context, cfg *model.Cfg, tree *lctree.Service, store *store.Service, log *logger.Log) (*Service, error) {
	service := &Service{
		cfg:             cfg,
		log:             log,
		store:           store,
		wg:              sync.WaitGroup{},
		updateTicker:    *time.NewTicker(24 * time.Hour),
		RPSLRouterClass: make(rpsl.RouterClass),
		tree:            tree,
	}

	var err error

	log.Info("Starting")

	service.radb, err = rpslsource.New(ctx, store, cfg, log.New("radb"), rpslsource.Config{
		Name:      "radb",
		Transport: rpslsource.TransportFTP,
		RemoteFiles: []rpslsource.RemoteFile{
			{Name: "radb", Path: "/radb/dbase/radb.db.gz"},
		},
		SerialPath:   "/radb/dbase/RADB.CURRENTSERIAL",
		Host:     "ftp.radb.net:21",
		FilePath: cfg.IPService.Radb.FilePath,
	})
	if err != nil {
		return nil, err
	}

	service.ripe, err = rpslsource.New(ctx, store, cfg, log.New("ripe"), rpslsource.Config{
		Name:      "ripe",
		Transport: rpslsource.TransportHTTP,
		RemoteFiles: []rpslsource.RemoteFile{
			{Name: "route6", Path: "/ripe/dbase/split/ripe.db.route6.gz"},
			{Name: "route", Path: "/ripe/dbase/split/ripe.db.route.gz"},
		},
		SerialPath:   "/ripe/dbase/RIPE.CURRENTSERIAL",
		Host:         "https://ftp.ripe.net",
		FilePath:     cfg.IPService.RIPE.FilePath,
		AddEOFMarker: true,
	})
	if err != nil {
		return nil, err
	}

	_, err = service.radb.Update(ctx)
	if err != nil {
		service.log.Error(err, "Error updating radb")
	}

	_, err = service.ripe.Update(ctx)
	if err != nil {
		service.log.Error(err, "Error updating ripe")
	}

	service.RPSLRouterClass, err = rpsl.RouterClassOpinionatedMerge(ctx, service.radb.RPSLRouterClass, service.ripe.RPSLRouterClass)
	if err != nil {
		return nil, err
	}
	service.radb.RPSLRouterClass = nil
	service.ripe.RPSLRouterClass = nil

	if err := service.tree.Build(ctx, service.RPSLRouterClass); err != nil {
		return nil, err
	}

	log.Info("Started")

	go func() {
		for {
			select {
			case <-service.updateTicker.C:
				log.Info("Starting whois update")
				radbUpdated, err := service.radb.Update(ctx)
				if err != nil {
					service.log.Error(err, "Error updating radb")
				}

				ripeUpdated, err := service.ripe.Update(ctx)
				if err != nil {
					service.log.Error(err, "Error updating ripe")
				}

				if radbUpdated && ripeUpdated {
					service.mu.Lock()
					service.RPSLRouterClass, err = rpsl.RouterClassOpinionatedMerge(ctx, service.radb.RPSLRouterClass, service.ripe.RPSLRouterClass)
					if err != nil {
						service.log.Error(err, "Error merging RPSL router classes")
					}
					service.radb.RPSLRouterClass = nil
					service.ripe.RPSLRouterClass = nil
					service.mu.Unlock()

					if err := service.tree.Build(ctx, service.RPSLRouterClass); err != nil {
						service.log.Error(err, "Error rebuilding patricia tree")
					}

					q, err := service.QueryIP(ctx, "2001:67c:2564::1")
					if err != nil {
						service.log.Error(err, "Error querying IP after update")
					}
					service.log.Info("Example query after update", "result", q)
				}
			case <-ctx.Done():
				service.log.Info("Stopping whois update")
				return
			}
		}
	}()

	return service, nil
}

func (s *Service) Close(ctx context.Context) error {
	s.log.Info("Quit")
	ctx.Done()

	if err := s.ripe.Close(ctx); err != nil {
		s.log.Error(err, "Error closing ripe service")
	}

	if err := s.radb.Close(ctx); err != nil {
		s.log.Error(err, "Error closing radb service")
	}

	return nil
}

// NewTestService creates a minimal Service for unit testing with a pre-built tree and router class.
func NewTestService(tree *lctree.Service, routerClass rpsl.RouterClass) *Service {
	return &Service{
		RPSLRouterClass: routerClass,
		tree:            tree,
	}
}
