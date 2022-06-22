package apiv1

import (
	"context"
	"ip_service/internal/maxmind"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
)

// Client holds the public api object
type Client struct {
	config *model.Cfg
	logger *logger.Logger
	max    *maxmind.Service
}

// New creates a new instance of public api
func New(ctx context.Context, max *maxmind.Service, config *model.Cfg, logger *logger.Logger) (*Client, error) {
	c := &Client{
		config: config,
		logger: logger,
		max:    max,
	}

	c.logger.Info("Started")

	return c, nil
}
