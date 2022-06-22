package configuration

import (
	"errors"
	"ip_service/pkg/helpers"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type envVars struct {
	ConfigYAML string `envconfig:"CONFIG_YAML" required:"true"`
}

// Parse parses config file from CONFIG_YAML environment variable
func Parse(logger *logger.Logger) (*model.Cfg, error) {
	logger.Info("Read environmental variable")
	var env envVars
	if err := envconfig.Process("", &env); err != nil {
		return nil, err
	}

	configPath := env.ConfigYAML

	config := &model.Config{}

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, errors.New("config is a folder")
	}

	if err := yaml.Unmarshal(configFile, config); err != nil {
		return nil, err
	}

	if err := helpers.Check(config, logger); err != nil {
		return nil, err
	}

	return config.IPService, nil
}
