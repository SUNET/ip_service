package httpserver

import (
	"context"
	"fmt"
	"ip_service/internal/apiv1"
	"ip_service/pkg/helpers"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Service is the service object for httpserver
type Service struct {
	config *model.Cfg
	logger *logger.Logger
	server *http.Server
	apiv1  Apiv1
	gin    *gin.Engine
}

// New creates a new httpserver service
func New(ctx context.Context, config *model.Cfg, api *apiv1.Client, logger *logger.Logger) (*Service, error) {
	s := &Service{
		config: config,
		logger: logger,
		apiv1:  api,
		server: &http.Server{Addr: config.APIServer.Host},
	}

	switch s.config.Production {
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

	// Middlewares
	s.gin.Use(s.middlewareDuration(ctx))
	s.gin.Use(s.middlewareLogger(ctx))
	s.gin.Use(s.middlewareCrash(ctx))
	s.gin.NoRoute(func(c *gin.Context) {
		status := http.StatusNotFound
		p := helpers.Problem404()
		c.JSON(status, gin.H{"error": p, "data": nil})
	})

	s.regEndpoint(ctx, "GET", "/", s.endpointIP)
	s.regEndpoint(ctx, "GET", "/city", s.endpointCity)
	s.regEndpoint(ctx, "GET", "/asn", s.endpointASN)
	s.regEndpoint(ctx, "GET", "/country", s.endpointCountry)
	s.regEndpoint(ctx, "GET", "/country-iso", s.endpointCountryISO)
	s.regEndpoint(ctx, "GET", "/json", s.endpointJSON)

	s.regEndpoint(ctx, "GET", "/lookup/:ip", s.endpointLookUpIP)

	s.regEndpoint(ctx, "GET", "/info", s.endpointInfo)

	s.regEndpoint(ctx, "GET", "/health", s.endpointStatus)

	// Run http server
	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			s.logger.New("http").Fatal("listen_error", "error", err)
		}
	}()

	s.logger.Info("started")

	return s, nil
}

func (s *Service) regEndpoint(ctx context.Context, method, path string, handler func(context.Context, *gin.Context) (interface{}, error)) {
	s.gin.Handle(method, path, func(c *gin.Context) {
		res, err := handler(ctx, c)
		if err != nil {
			c.IndentedJSON(400, gin.H{"data": nil, "error": helpers.NewErrorFromError(err)})
		}

		switch c.Request.Header.Get("Accept") {
		case "application/json":
			c.IndentedJSON(200, res)
		case "text/plain":
			c.Writer.WriteString(fmt.Sprintf("%v", res))
		case "*/*":
			c.Writer.WriteString(fmt.Sprintf("%v\n", res))

		}
	})
}

// Close closing httpserver
func (s *Service) Close(ctx context.Context) error {
	s.logger.Info("Quit")
	return nil
}
