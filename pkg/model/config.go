package model

import "time"

// APIServer holds the api server configuration
type APIServer struct {
	Addr string `yaml:"addr" validate:"required"`
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
	LicenseKey        string               `yaml:"license_key"`
	Enterprise        bool                 `yaml:"enterprise"`
	RetryCounter      int                  `yaml:"retry_counter"`
	DB                map[string]MaxMindDB `yaml:"db" validate:"required"`
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
	Addr string `yaml:"addr"`
}

// IPService configs ip_service
type IPService struct {
	APIServer  APIServer `yaml:"api_server"`
	Production bool      `yaml:"production"`
	HTTPProxy  string    `yaml:"http_proxy"`
	Log        Log       `yaml:"log"`
	MaxMind    MaxMind   `yaml:"maxmind" validate:"required"`
	Store      Store     `yaml:"store"`
	Tracing    Tracing   `yaml:"tracing"`
}

// Cfg holds the configuration for the service
type Cfg struct {
	IPService *IPService `yaml:"ip_service" validate:"required"`
}
