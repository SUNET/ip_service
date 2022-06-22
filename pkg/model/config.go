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

	Redis struct {
		DB                  int      `yaml:"db" validate:"required"`
		Addr                string   `yaml:"host" validate:"required_without_all=SentinelHosts SentinelServiceName"`
		SentinelHosts       []string `yaml:"sentinel_hosts" validate:"required_without=Addr,omitempty,min=2,max=4"`
		SentinelServiceName string   `yaml:"sentinel_service_name" validate:"required_with=SentinelHosts"`
	} `yaml:"redis"`
}

type Config struct {
	IPService *Cfg `yaml:"ip_service"`
}
