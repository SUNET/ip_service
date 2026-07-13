package maxmind

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"ip_service/internal/store"
	"ip_service/pkg/helpers"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/model"
	"github.com/SUNET/vc/pkg/trace"

	"github.com/oschwald/geoip2-golang"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/time/rate"
)

type DBMeta map[string]*DBObject

func (d DBMeta) DownloadInProgress(dbType string) {
	d[dbType].Downloading = true
}

func (d DBMeta) DownloadingDone(dbType string) {
	d[dbType].Downloading = false
}

func (d DBMeta) IsDownloadingInProgresses(dbType string) bool {
	return d[dbType].Downloading
}

// Service holds the maxmind service object
type Service struct {
	probeStore   *model.StatusProbeStore
	cfg          *model.Cfg
	Log          *logger.Log
	TP           *trace.Tracer
	httpClient   *http.Client
	DBMeta       DBMeta
	DBCity       *geoip2.Reader
	DBASN        *geoip2.Reader
	kvStore      kvStore
	quitChan     chan struct{}
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
	rateLimit   rate.Limiter
	MU          sync.RWMutex
	Missing     bool
	Downloading bool
}

// New creates a new instance of maxmind
func New(ctx context.Context, cfg *model.Cfg, store *store.Service, tp *trace.Tracer, log *logger.Log) (*Service, error) {
	s := &Service{
		probeStore:   &model.StatusProbeStore{},
		cfg:          cfg,
		Log:          log,
		TP:           tp,
		kvStore:      store.KV,
		quitChan:     make(chan struct{}, 1),
		reloadChan:   make(chan string, 10),
		downloadChan: make(chan string, 10),
		updateChan:   make(chan string, 10),
		initialChan:  make(chan string, 10),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		DBMeta: map[string]*DBObject{
			model.MaxmindDBTypeCity: {
				Missing:     true,
				Downloading: false,
				rateLimit:   *rate.NewLimiter(rate.Every(24*time.Hour), 4),
			},
			model.MaxmindDBTypeASN: {
				Missing:     true,
				Downloading: false,
				rateLimit:   *rate.NewLimiter(rate.Every(24*time.Hour), 4),
			},
		},
	}

	updateTicker := time.NewTicker(s.cfg.IPService.MaxMind.UpdatePeriodicity * time.Second)

	if err := os.MkdirAll(s.cfg.IPService.MaxMind.BaseFolder, 0750); err != nil {
		return nil, err
	}

	for dbType := range s.DBMeta {
		s.Log.Info("init db", "dbType", dbType)
		if err := s.initial(ctx, dbType); err != nil {
			s.Log.Error(err, "failed to initialize db", "dbType", dbType)
			return nil, err
		}
	}

	s.Log.Info("Started")

	go func() {
		for {
			select {
			// check periodically if there is a new version of the database available
			case <-updateTicker.C:
				s.Log.Info("UpdateTicker")
				if cfg.IPService.MaxMind.AutomaticUpdate {
					for dbType := range s.DBMeta {
						if !s.DBMeta.IsDownloadingInProgresses(dbType) {
							s.Log.Info("No downloading in progress", "dbtype", dbType)
							s.updateChan <- dbType
						}
					}
				}

			case dbType := <-s.updateChan:
				if s.checkNewDBVersion(ctx, dbType) || !s.cfg.IPService.MaxMind.IsDBPresent(dbType) {
					s.Log.Debug("updateChan", "dbType", dbType)
					s.downloadChan <- dbType
				}

			case dbType := <-s.downloadChan:
				s.Log.Info("downloadChan", "dbType", dbType)
				if err := s.downloadArchive(ctx, dbType); err != nil {
					s.Log.Error(err, "dbDownloader")
					return
				}

			case dbType := <-s.reloadChan:
				s.Log.Info("reloadChan", "dbType", dbType)
				if err := s.loadDB(ctx, dbType); err != nil {
					s.Log.Error(err, "loadDB")
				}

			case <-s.quitChan:
				s.Log.Info("quit database update")
				updateTicker.Stop()
				return
			}
		}
	}()

	return s, nil
}

func (s *Service) loadDB(ctx context.Context, dbType string) error {
	s.DBMeta[dbType].MU.Lock()
	defer s.DBMeta[dbType].MU.Unlock()

	_, span := s.TP.Start(ctx, "maxmind:loadDB")
	defer span.End()

	s.Log.Info("Run loadDB for", "dbType", dbType)

	if !s.cfg.IPService.MaxMind.IsDBPresent(dbType) {
		return helpers.ErrMissingDBFile
	}

	dbFileName := s.cfg.IPService.MaxMind.DBFilePath(dbType)

	s.Log.Debug("load", "dbFileName", dbFileName)

	db, err := geoip2.Open(dbFileName)
	if err != nil {
		s.Log.Error(err, "geoip2.Open failed")
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if db == nil {
		return errors.New("geoip2.Open returned nil db")
	}

	switch dbType {
	case model.MaxmindDBTypeCity:
		s.DBCity = db
	case model.MaxmindDBTypeASN:
		s.DBASN = db
		metadata := s.DBASN.Metadata()
		s.Log.Info("Maxmind", "dbType", dbType, "metadata", metadata)
	default:
		err := errors.New("unknown dbType: " + dbType)
		s.Log.Error(err, "cannot load db")
		return err
	}

	return nil
}

// Close closes maxmind service
func (s *Service) Close(ctx context.Context) error {
	s.Log.Info("Quit")
	ctx.Done()
	return nil
}
