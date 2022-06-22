package maxmind

import (
	"context"
	"ip_service/internal/store"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"path/filepath"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// Service holds the maxmind service object
type Service struct {
	cfg      *model.Cfg
	log      *logger.Logger
	dbFiles  map[string]dbObject
	dbCity   *geoip2.Reader
	dbASN    *geoip2.Reader
	kvStore  kvStore
	quitChan chan bool
}

type kvStore interface {
	Set(ctx context.Context, k string, v interface{}) error
	Get(ctx context.Context, k string) (string, error)
}

type dbObject struct {
	url      string
	filePath string
}

// New creates a new instance of maxmind
func New(ctx context.Context, cfg *model.Cfg, store *store.Service, log *logger.Logger) (*Service, error) {
	s := &Service{
		cfg:      cfg,
		log:      log,
		kvStore:  store.KV,
		quitChan: make(chan bool),
		dbFiles: map[string]dbObject{
			"City": {
				filePath: filepath.Join("db", "GeoLite2-City.mmdb"),
				url:      "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz",
			},
			"ASN": {
				filePath: filepath.Join("db", "GeoLite2-ASN.mmdb"),
				url:      "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=%s&suffix=tar.gz",
			},
		},
	}

	for dbType, object := range s.dbFiles {
		s.openDB(ctx, dbType, object.url, object.filePath)
	}


	ticker := time.NewTicker(time.Duration(s.cfg.MaxMind.UpdatePeriodicity * time.Second))
	go func() {
		for {
			select {
			case <-ticker.C:
				for dbType, object := range s.dbFiles {
					s.openDB(ctx, dbType, object.url, object.filePath)
				}
			case <-s.quitChan:
				s.log.Info("quit database update")
				ticker.Stop()
				return
			}
		}
	}()

	return s, nil
}

func (s *Service) openDB(ctx context.Context, dbType, url, filePath string) {
	s.log.Info("Run openDB for", "dbType", dbType)
	haveNewVersion, err := s.isNewVersion(ctx, dbType, url)
	if err != nil {
		s.log.Warn("Error", "value", err)
	}
	if haveNewVersion {
		if err := s.fetchDatabase(ctx, url, dbType); err != nil {
			s.log.Warn("Error", "value", err)
		}

		if err := s.unTAR(dbType); err != nil {
			s.log.Warn("Error", "value", err)
		}

		if err := s.cleanUpTarArchive(dbType); err != nil {
			s.log.Warn("Error", "value", err)
		}
	}

	switch dbType {
	case "City":
		db, err := geoip2.Open(filePath)
		if err != nil {
			s.log.Warn("Error", "value", err)
		}
		s.dbCity = db
	case "ASN":
		db, err := geoip2.Open(filePath)
		if err != nil {
			s.log.Warn("Error", "value", err)
		}
		s.dbASN = db
	}
}

// Close closes maxmind service
func (s *Service) Close(ctx context.Context) error {
	s.quitChan <- true
	ctx.Done()
	return nil
}
