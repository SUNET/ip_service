package maxmind

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"ip_service/internal/store"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"

	"github.com/oschwald/geoip2-golang"
)

// Service holds the maxmind service object
type Service struct {
	cfg      *model.Cfg
	log      *logger.Logger
	dbMeta   map[string]dbObject
	DBCity   *geoip2.Reader
	DBASN    *geoip2.Reader
	kvStore  kvStore
	quitChan chan bool
}

type kvStore interface {
	GetRemoteVersion(ctx context.Context, k string) string
	SetLastChecked(ctx context.Context, k string) error
	SetPreviousVersion(ctx context.Context, k string) error
	SetRemoteVersion(ctx context.Context, k, v string) error
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
		dbMeta: map[string]dbObject{
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

	for dbType := range s.dbMeta {
		s.openDB(ctx, dbType)
	}

	ticker := time.NewTicker(s.cfg.MaxMind.UpdatePeriodicity * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if cfg.MaxMind.AutomaticUpdate {
					for dbType := range s.dbMeta {
						if err := s.openDB(ctx, dbType); err != nil {
							s.log.Warn("Error", "value", err)
						}
					}
					if !s.Status(ctx).Healthy {
						s.log.Warn("kind", "value", reflect.TypeOf(s.Status(ctx).ErrorType))
					}
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

func (s *Service) openDB(ctx context.Context, dbType string) error {
	var missingDBFile bool

	s.kvStore.SetLastChecked(ctx, dbType)

	s.log.Info("Run openDB for", "dbType", dbType)
	if _, err := os.Stat("db"); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir("db", os.ModePerm); err != nil {
			return err
		}
	}

	if _, err := os.Stat(s.dbMeta[dbType].filePath); errors.Is(err, os.ErrNotExist) {
		missingDBFile = true
	}

	haveNewVersion, err := s.isNewVersion(ctx, dbType)
	if err != nil {
		return err
	}

	if haveNewVersion || missingDBFile {
		s.log.Info("Getting the latest remote database", "dbType", dbType)
		if err := s.getLatestDB(ctx, dbType); err != nil {
			return err
		}

		s.log.Info("UnTar", "dbType", dbType)
		if err := s.unTAR(dbType); err != nil {
			return err
		}

		s.log.Info("Cleanup tar archive for database", "dbType", dbType)
		if err := s.cleanUpTarArchive(dbType); err != nil {
			return err
		}
	}

	db, err := geoip2.Open(s.dbMeta[dbType].filePath)
	if err != nil {
		return err
	}

	switch dbType {
	case "City":
		s.DBCity = db
	case "ASN":
		s.DBASN = db
	}

	return nil
}

var (
	nextRun    time.Time
	lastStatus *model.StatusService
)

func (s *Service) Status(ctx context.Context) *model.StatusService {
	if time.Now().After(nextRun) {
		status := &model.StatusService{
			ServiceName: "maxmind",
			Timestamp:   time.Now(),
			Interval:    10 * time.Second,
		}

		for _, testIP := range []string{"95.142.107.181", "110.50.243.6", "69.162.81.155"} {
			_, err := s.DBCity.Country(net.ParseIP(testIP))
			if err != nil {
				status.Message = fmt.Sprintf("dbCity country: %v", err)
				s.log.Warn("type", "value", reflect.TypeOf(err))
				return status
			}

			_, err = s.DBASN.ASN(net.ParseIP(testIP))
			if err != nil {
				status.Message = fmt.Sprintf("dbASN ASN %v", err)
				s.log.Warn("type", "value", reflect.TypeOf(err))
				return status
			}
		}

		status.Healthy = true
		nextRun = time.Now().Add(status.Interval)
		lastStatus = status

		return status
	}

	return lastStatus
}

// Close closes maxmind service
func (s *Service) Close(ctx context.Context) error {
	s.log.Info("Quit")
	s.quitChan <- true
	ctx.Done()
	return nil
}
