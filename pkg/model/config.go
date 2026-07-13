package model

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// APIServer holds the api server configuration
type APIServer struct {
	Addr        string `yaml:"addr" validate:"required"`
	BehindProxy bool   `yaml:"behind_proxy"`
}

// Log holds the log configuration
type Log struct {
	Level      string `yaml:"level"`
	FolderPath string `yaml:"folder_path"`
}

// MaxMindDB holds maxmind db configuration
type MaxMindDB struct {
	FilePath string `yaml:"file_path"`
}

// MaxMind holds the maxmind configuration
type MaxMind struct {
	AutomaticUpdate   bool                 `yaml:"automatic_update"`
	UpdatePeriodicity time.Duration        `yaml:"update_periodicity"`
	BaseFolder        string               `yaml:"base_folder" validate:"required"`
	Username          string               `yaml:"username" validate:"required"`
	Password          string               `yaml:"password" validate:"required"`
	Enterprise        bool                 `yaml:"enterprise"`
	RetryCounter      int                  `yaml:"retry_counter"`
	DB                map[string]MaxMindDB `yaml:"db" validate:"required"`
	RemoteURL         string               `yaml:"remote_url" validate:"required"`
	ArchiveFormat     string               `yaml:"archive_format" validate:"required,oneof=tar.gz"`
}

func (m *MaxMind) URL(dbType string) (string, error) {
	u, err := url.Parse(m.RemoteURL)
	if err != nil {
		return "", err
	}

	u = u.JoinPath("GeoLite2-" + dbType + "/download")

	q := u.Query()
	q.Set("suffix", m.ArchiveFormat)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (m *MaxMind) IsArchivePresent(dbType string) bool {
	p := filepath.Join(m.BaseFolder, fmt.Sprintf("GeoLite2-%s.tar.gz", dbType))

	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		return true
	}
	return false
}

func (m *MaxMind) IsDBPresent(dbType string) bool {
	databasePath := m.DBFilePath(dbType)

	_, err := os.Stat(databasePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		return false
	}

	return true
}

func (m *MaxMind) ArchiveFilePath(dbType string) string {
	return filepath.Join(m.BaseFolder, fmt.Sprintf("GeoLite2-%s.tar.gz", dbType))
}

// DBFilePath returns the file path for the given dbType (e.g. <basefolder>/GeoLite2-ASN.mmdb)
func (m *MaxMind) DBFilePath(dbType string) string {
	return filepath.Join(m.BaseFolder, fmt.Sprintf("GeoLite2-%s.mmdb", dbType))
}

type Radb struct {
	FilePath string `yaml:"file_path"`
}

type RIPE struct {
	FilePath string `yaml:"file_path"`
}

// FileStorage holds the file storage configuration
type FileStorage struct {
	Path string `yaml:"path"`
}

// Store holds the store configuration
type Store struct {
	File FileStorage `yaml:"file"`
}

// Tracing holds the tracing configuration
type Tracing struct {
	Addr   string `yaml:"addr"`
	Enable bool   `yaml:"enable"`
}

// IPService configs ip_service
type IPService struct {
	APIServer  APIServer `yaml:"api_server"`
	Production bool      `yaml:"production"`
	Log        Log       `yaml:"log"`
	MaxMind    MaxMind   `yaml:"maxmind" validate:"required"`
	Radb       Radb      `yaml:"radb" validate:"required"`
	RIPE       RIPE      `yaml:"ripe" validate:"required"`
	Store      Store     `yaml:"store"`
	Tracing    Tracing   `yaml:"tracing"`
}

// Cfg holds the configuration for the service
type Cfg struct {
	IPService *IPService `yaml:"ip_service" validate:"required"`
}
