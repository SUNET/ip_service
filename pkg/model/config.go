package model

import "time"

// Cfg is the main configuration structure for this application
type Cfg struct {
	APIServer struct {
		Host string `yaml:"host" validate:"required"`
	} `yaml:"api_server"`

	Production bool   `yaml:"production"`
	HTTPProxy  string `yaml:"http_proxy"`

	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`

	MaxMind struct {
		UpdatePeriodicity time.Duration `yaml:"update_periodicity"`
		LicenseKey        string        `yaml:"license_key"`
		Enterprise        bool          `yaml:"enterprise"`
	} `yaml:"max_mind"`

	Store struct {
		File struct {
			Path string `yaml:"path"`
		} `yaml:"file"`
	} `yaml:"store"`
}

type Config struct {
	IPService *Cfg `yaml:"ip_service"`
}
