package apiv1

import (
	"context"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/internal/whois"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/model"
	"github.com/SUNET/vc/pkg/trace"
)

//	@title		ip_service API
//	@version	0.1.0
//	@BasePath	/

// Client holds the public api object
type Client struct {
	config *model.Cfg
	log    *logger.Log
	tp     *trace.Tracer
	max    *maxmind.Service
	whois  *whois.Service
	store  *store.Service
}

// New creates a new instance of public api
func New(ctx context.Context, max *maxmind.Service, whois *whois.Service, store *store.Service, config *model.Cfg, tp *trace.Tracer, log *logger.Log) (*Client, error) {
	c := &Client{
		config: config,
		log:    log,
		tp:     tp,
		max:    max,
		whois:  whois,
		store:  store,
	}

	c.log.Info("Started")

	return c, nil
}
