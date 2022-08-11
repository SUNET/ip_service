package apiv1

import (
	"context"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
)

// Client holds the public api object
type Client struct {
	config *model.Cfg
	logger *logger.Logger
	max    *maxmind.Service
	store  *store.Service
}

// New creates a new instance of public api
func New(ctx context.Context, max *maxmind.Service, store *store.Service, config *model.Cfg, logger *logger.Logger) (*Client, error) {
	c := &Client{
		config: config,
		logger: logger,
		max:    max,
		store:  store,
	}

	c.logger.Info("Started")

	return c, nil
}
