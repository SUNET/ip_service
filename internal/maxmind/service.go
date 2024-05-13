package maxmind

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"ip_service/internal/store"
	"ip_service/pkg/helpers"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/trace"

	"github.com/oschwald/geoip2-golang"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/time/rate"
)

// Service holds the maxmind service object
type Service struct {
	probeStore   *model.StatusProbeStore
	wg           sync.WaitGroup
	cfg          *model.Cfg
	log          *logger.Log
	TP           *trace.Tracer
	DBMeta       map[string]*DBObject
	DBCity       *geoip2.Reader
	DBASN        *geoip2.Reader
	kvStore      kvStore
	quitChan     chan bool
	reloadChan   chan string
	downloadChan chan string
	updateChan   chan string
	initialChan  chan string
}

type kvStore interface {
	GetRemoteVersion(ctx context.Context, k string) string
	GetLastChecked(ctx context.Context, k string) string
	SetLastChecked(ctx context.Context, k string) error
	SetPreviousVersion(ctx context.Context, k string) error
	SetRemoteVersion(ctx context.Context, k, v string) error
}

// DBObject holds the maxmind database object
type DBObject struct {
	urlDB     string
	filePath  string
	rateLimit rate.Limiter
	MU        sync.RWMutex
}

// New creates a new instance of maxmind
func New(ctx context.Context, cfg *model.Cfg, store *store.Service, tp *trace.Tracer, log *logger.Log) (*Service, error) {
	s := &Service{
		probeStore:   &model.StatusProbeStore{},
		wg:           sync.WaitGroup{},
		cfg:          cfg,
		log:          log,
		TP:           tp,
		kvStore:      store.KV,
		quitChan:     make(chan bool),
		reloadChan:   make(chan string, 10),
		downloadChan: make(chan string, 10),
		updateChan:   make(chan string, 10),
		initialChan:  make(chan string, 10),
		DBMeta: map[string]*DBObject{
			"city": {
				filePath:  cfg.IPService.MaxMind.DB["city"].FilePath,
				urlDB:     "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz",
				rateLimit: *rate.NewLimiter(rate.Every(24*time.Hour), 4),
			},
			"asn": {
				filePath:  cfg.IPService.MaxMind.DB["asn"].FilePath,
				urlDB:     "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=%s&suffix=tar.gz",
				rateLimit: *rate.NewLimiter(rate.Every(24*time.Hour), 4),
			},
		},
	}

	updateTicker := time.NewTicker(s.cfg.IPService.MaxMind.UpdatePeriodicity * time.Second)

	s.wg.Add(1)
	go func() {
		for {
			select {
			// check periodically if there is a new version of the database available
			case <-updateTicker.C:
				if cfg.IPService.MaxMind.AutomaticUpdate {
					for dbType := range s.DBMeta {
						s.updateChan <- dbType
					}
				}

			case dbType := <-s.initialChan:
				s.log.Info("initialChan", "dbType", dbType)
				if err := s.initial(ctx, dbType); err != nil {
					s.log.Error(err, "initial")
				}

			case dbType := <-s.updateChan:
				if s.checkNewDBVersion(ctx, dbType) || s.dbFilesMissing(ctx, dbType) {
					s.log.Debug("updateChan", "dbType", dbType)
					s.downloadChan <- dbType
				}

			case dbType := <-s.downloadChan:
				s.log.Info("downloadChan", "dbType", dbType)
				if err := s.dbDownloader(ctx, dbType); err != nil {
					s.log.Error(err, "dbDownloader")
				}
				s.reloadChan <- dbType

			case dbType := <-s.reloadChan:
				s.log.Info("reloadChan", "dbType", dbType)
				if err := s.loadDB(ctx, dbType); err != nil {
					s.log.Error(err, "loadDB")
				}

			case <-s.quitChan:
				s.log.Info("quit database update")
				updateTicker.Stop()
				s.wg.Done()
				return
			}
		}
	}()

	for dbType := range s.DBMeta {
		s.log.Debug("init db", "dbType", dbType)
		s.initialChan <- dbType
	}

	return s, nil
}

func (s *Service) loadDB(ctx context.Context, dbType string) error {
	s.DBMeta[dbType].MU.Lock()
	defer s.DBMeta[dbType].MU.Unlock()

	_, span := s.TP.Start(ctx, "maxmind:loadDB")
	defer span.End()

	s.log.Info("Run loadDB for", "dbType", dbType)

	if _, err := os.Stat(s.DBMeta[dbType].filePath); errors.Is(err, os.ErrNotExist) {
		s.log.Error(err, "loadDB", "dbType", dbType, "path", s.DBMeta[dbType].filePath)
		return helpers.ErrMissingDBFile
	}

	db, err := geoip2.Open(s.DBMeta[dbType].filePath)
	if err != nil {
		s.log.Error(err, "geoip2.Open failed")
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	switch dbType {
	case "city":
		s.DBCity = db
	case "asn":
		s.DBASN = db
	}

	return nil
}

// Close closes maxmind service
func (s *Service) Close(ctx context.Context) error {
	s.quitChan <- true
	s.wg.Wait()
	ctx.Done()
	s.log.Info("Quit")
	return nil
}
