package apiv1

import (
	"context"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/trace"
)

// Client holds the public api object
type Client struct {
	config *model.Cfg
	log    *logger.Log
	tp     *trace.Tracer
	max    *maxmind.Service
	store  *store.Service
}

// New creates a new instance of public api
func New(ctx context.Context, max *maxmind.Service, store *store.Service, config *model.Cfg, tp *trace.Tracer, log *logger.Log) (*Client, error) {
	c := &Client{
		config: config,
		log:    log,
		tp:     tp,
		max:    max,
		store:  store,
	}

	c.log.Info("Started")

	return c, nil
}
