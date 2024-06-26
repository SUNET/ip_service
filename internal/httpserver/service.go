package httpserver

import (
	"context"
	"fmt"
	"ip_service/internal/apiv1"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/helpers"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/trace"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Service is the service object for httpserver
type Service struct {
	config  *model.Cfg
	logger  *logger.Log
	TP      *trace.Tracer
	metrics *metrics
	server  *http.Server
	apiv1   Apiv1
	gin     *gin.Engine
}

// New creates a new httpserver service
func New(ctx context.Context, cfg *model.Cfg, api *apiv1.Client, tp *trace.Tracer, logger *logger.Log) (*Service, error) {
	s := &Service{
		config:  cfg,
		logger:  logger,
		TP:      tp,
		metrics: &metrics{},
		apiv1:   api,
		server: &http.Server{
			Addr:              cfg.IPService.APIServer.Addr,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	s.metrics.init()

	switch s.config.IPService.Production {
	case true:
		gin.SetMode(gin.ReleaseMode)
	case false:
		gin.SetMode(gin.DebugMode)
	}

	apiValidator := validator.New()
	binding.Validator = &defaultValidator{
		Validate: apiValidator,
	}

	s.gin = gin.New()
	s.server.Handler = s.gin
	s.server.ReadTimeout = time.Second * 5
	s.server.WriteTimeout = time.Second * 30
	s.server.IdleTimeout = time.Second * 90

	s.gin.LoadHTMLGlob("templates/*.html")
	s.gin.Static("/assets", "./assets")

	// Middlewares
	s.gin.Use(s.middlewareDuration(ctx))
	s.gin.Use(s.middlewareTraceID(ctx))
	s.gin.Use(s.middlewareLogger(ctx))
	s.gin.Use(s.middlewareCrash(ctx))
	s.gin.NoRoute(func(c *gin.Context) {
		status := http.StatusNotFound
		p, err := helpers.Problem404()
		if err != nil {
			c.JSON(status, gin.H{"error": err, "data": nil})
		}
		c.JSON(status, gin.H{"error": p, "data": nil})
	})

	rgRoot := s.gin.Group("/")

	s.regEndpoint(ctx, "GET", "/", s.endpointIndex)
	s.regEndpoint(ctx, "GET", "/city", s.endpointCity)
	s.regEndpoint(ctx, "GET", "/asn", s.endpointASN)
	s.regEndpoint(ctx, "GET", "/country", s.endpointCountry)
	s.regEndpoint(ctx, "GET", "/country-iso", s.endpointCountryISO)
	s.regEndpoint(ctx, "GET", "/coordinates", s.endpointCoordinates)
	s.regEndpoint(ctx, "GET", "/all", s.endpointAll)

	s.regEndpoint(ctx, "GET", "/lookup/:ip", s.endpointLookUpIP)

	s.regEndpoint(ctx, "GET", "/health", s.endpointHealth)

	rgMetrics := rgRoot.Group("/metrics")
	rgMetrics.GET("/", gin.WrapH(promhttp.Handler()))

	// Run http server
	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			s.logger.New("http").Error(err, "listen_error")
		}
	}()

	s.logger.Info("started")

	return s, nil
}

// client: Accept -> server
// server: content-type -> client
func (s *Service) getAccept(c *gin.Context) string {
	switch c.Request.Header.Get("Accept") {
	case gin.MIMEHTML:
		return gin.MIMEHTML
	case gin.MIMEJSON:
		return gin.MIMEJSON
	case gin.MIMEPlain:
		return gin.MIMEPlain
	case "*/*":
		return gin.MIMEPlain
	default:
		return gin.MIMEHTML
	}
}

func (s *Service) regEndpoint(ctx context.Context, method, path string, handler func(context.Context, *gin.Context) (any, error)) {
	s.gin.Handle(method, path, func(c *gin.Context) {
		ctx = contexthandler.Add(ctx, "request", &contexthandler.RequestContext{
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Accept:    s.getAccept(c),
		})
		res, err := handler(ctx, c)
		if err != nil {
			c.IndentedJSON(400, gin.H{"data": nil, "error": helpers.NewErrorFromError(err)})
		}

		requestValues, err := contexthandler.Get(ctx, "request")
		if err != nil {
			c.IndentedJSON(400, gin.H{"data": nil, "error": helpers.NewErrorFromError(err)})
		}

		switch requestValues.Accept {
		case gin.MIMEHTML:
			switch r := res.(type) {
			case *model.ReplyIPInformation:
				c.HTML(http.StatusOK, "index.html", r)
			}
		case gin.MIMEJSON:
			c.IndentedJSON(200, res)
		case gin.MIMEPlain:
			_, err := c.Writer.WriteString(fmt.Sprintf("%v\n", res))
			if err != nil {
				c.IndentedJSON(400, gin.H{"data": nil, "error": helpers.NewErrorFromError(err)})
			}
		}
	})
}

// Close closing httpserver
func (s *Service) Close(ctx context.Context) error {
	s.logger.Info("Quit")
	return nil
}
