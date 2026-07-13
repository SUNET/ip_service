package rpslsource

import (
	"context"
	"ip_service/internal/store"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/rpsl"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

// TransportType defines the download transport
type TransportType int

const (
	TransportFTP  TransportType = iota
	TransportHTTP TransportType = iota
)

// RemoteFile defines a single remote archive to download
type RemoteFile struct {
	// Name identifies this archive (used for local file naming)
	Name string
	// Path is the remote FTP path or HTTP URL
	Path string
}

// Config defines how an RPSL source behaves
type Config struct {
	// Name is the identifier for this source (e.g. "radb", "ripe")
	Name string
	// Transport is FTP or HTTP
	Transport TransportType
	// RemoteFiles lists the remote archives to download and parse
	RemoteFiles []RemoteFile
	// SerialPath is the path (FTP) or URL (HTTP) for fetching the current serial
	SerialPath string
	// Host is the host:port for FTP connections (only for FTP transport)
	Host string
	// FilePath is the base path for persisted local files
	FilePath string
	// AddEOFMarker appends an EOF marker after unzip (needed for RIPE)
	AddEOFMarker bool
}

// Service is a generic RPSL source that handles downloading, caching, and parsing
type Service struct {
	cfg             *model.Cfg
	sourceCfg       Config
	log             *logger.Log
	RPSLRouterClass rpsl.RouterClass
	store           kvStore
	rateLimit       rate.Limiter
	httpClient      *http.Client
}

type kvStore interface {
	GetRemoteVersion(ctx context.Context, k string) string
	GetLastChecked(ctx context.Context, k string) string
	SetLastChecked(ctx context.Context, k string) error
	SetPreviousVersion(ctx context.Context, k string) error
	SetRemoteVersion(ctx context.Context, k, v string) error
}

// New creates a new RPSL source service
func New(ctx context.Context, store *store.Service, cfg *model.Cfg, log *logger.Log, sourceCfg Config) (*Service, error) {
	service := &Service{
		cfg:             cfg,
		sourceCfg:       sourceCfg,
		log:             log,
		RPSLRouterClass: make(rpsl.RouterClass),
		store:           store.KV,
		rateLimit:       *rate.NewLimiter(rate.Every(24*time.Hour), 4),
	}

	if sourceCfg.Transport == TransportHTTP {
		retHTTPClient := retryablehttp.NewClient()
		retHTTPClient.RetryMax = 10
		service.httpClient = retHTTPClient.StandardClient()
	}

	if err := os.MkdirAll(filepath.Dir(sourceCfg.FilePath), 0750); err != nil {
		return nil, err
	}

	return service, nil
}

// Update checks for updates and downloads/parses RPSL data. Uses local cache when possible.
func (s *Service) Update(ctx context.Context) (bool, error) {
	storeKey := s.sourceCfg.Name + "_serial"
	currentSerial := s.store.GetRemoteVersion(ctx, storeKey)
	s.log.Info("Current serial", "source", s.sourceCfg.Name, "serial", currentSerial)

	currentRemoteSerial, err := s.getRemoteSerial(ctx)
	if err != nil {
		return false, err
	}

	if currentSerial == currentRemoteSerial {
		if s.loadFromLocal(ctx) {
			s.log.Info("Loaded from local cache", "source", s.sourceCfg.Name, "serial", currentSerial)
			return true, nil
		}
		s.log.Info("Local cache missing, downloading despite matching serial", "source", s.sourceCfg.Name)
	} else {
		s.log.Info("New serial detected", "source", s.sourceCfg.Name, "serial", currentRemoteSerial)
	}

	if err := s.store.SetRemoteVersion(ctx, storeKey, currentRemoteSerial); err != nil {
		return false, err
	}

	rpslClient, err := rpsl.New(ctx)
	if err != nil {
		return false, err
	}

	for _, remoteFile := range s.sourceCfg.RemoteFiles {
		if err := s.downloadArchive(ctx, remoteFile); err != nil {
			return false, err
		}

		if err := s.unzip(ctx, remoteFile.Name); err != nil {
			return false, err
		}

		if s.sourceCfg.AddEOFMarker {
			if err := s.addEOFMarker(remoteFile.Name); err != nil {
				return false, err
			}
		}

		if err := s.cleanupArchive(remoteFile.Name); err != nil {
			return false, err
		}

		localPath := s.localFilePath(remoteFile.Name)
		if err := rpslClient.Parse(ctx, localPath); err != nil {
			return false, err
		}
	}
	s.RPSLRouterClass = rpslClient.RouterClass

	return true, nil
}

// Close shuts down the service
func (s *Service) Close(ctx context.Context) error {
	s.log.Info("Quit", "source", s.sourceCfg.Name)
	return nil
}
