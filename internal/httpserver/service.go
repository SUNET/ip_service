package httpserver

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"strings"

	"ip_service/internal/apiv1"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/helpers"
	"ip_service/pkg/model"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/template/html/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// Swagger
	_ "ip_service/docs"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

const (
	MIMEPlain = "text/plain"
	MIMEJSON  = "application/json"
	MIMEHTML  = "text/html"
)

// Service is the service object for httpserver
type Service struct {
	config  *model.Cfg
	logger  *logger.Log
	TP      *trace.Tracer
	metrics *metrics
	apiv1   Apiv1
	app     *fiber.App
}

// New creates a new httpserver service
func New(ctx context.Context, cfg *model.Cfg, api *apiv1.Client, tp *trace.Tracer, logger *logger.Log) (*Service, error) {
	tmplFS, _ := fs.Sub(templatesFS, "templates")
	engine := html.NewFileSystem(http.FS(tmplFS), ".html")

	s := &Service{
		config:  cfg,
		logger:  logger,
		TP:      tp,
		metrics: &metrics{},
		apiv1:   api,
	}

	s.metrics.init()

	s.app = fiber.New(fiber.Config{
		Views:                 engine,
		DisableStartupMessage: cfg.IPService.Production,
		ReadBufferSize:        8192,
		Prefork:               false,
		Network:               "tcp",
	})

	// Static files (embedded)
	s.app.Get("/assets/*", func(c *fiber.Ctx) error {
		fpath := c.Params("*")
		data, err := assetsFS.ReadFile("assets/" + fpath)
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		ctype := mime.TypeByExtension(filepath.Ext(fpath))
		if ctype != "" {
			c.Set("Content-Type", ctype)
		}
		return c.Send(data)
	})

	// Middlewares
	s.app.Use(s.middlewareDuration(ctx))
	s.app.Use(s.middlewareTraceID(ctx))
	s.app.Use(s.middlewareLogger(ctx))
	s.app.Use(s.middlewareCrash(ctx))

	// Swagger
	s.app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Endpoints
	s.regEndpoint(ctx, "GET", "/", s.endpointIndex)
	s.regEndpoint(ctx, "GET", "/city", s.endpointCity)
	s.regEndpoint(ctx, "GET", "/asn", s.endpointASN)
	s.regEndpoint(ctx, "GET", "/country", s.endpointCountry)
	s.regEndpoint(ctx, "GET", "/country-iso", s.endpointCountryISO)
	s.regEndpoint(ctx, "GET", "/coordinates", s.endpointCoordinates)
	s.regEndpoint(ctx, "GET", "/all", s.endpointAll)
	s.regEndpoint(ctx, "POST", "/collision", s.endpointCollision)

	s.regEndpoint(ctx, "GET", "/lookup/:ip", s.endpointLookUpIP)

	s.regEndpoint(ctx, "GET", "/whois/:ip", s.endpointWhois)

	s.regEndpoint(ctx, "GET", "/health", s.endpointHealth)

	// Metrics
	s.app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Debug/pprof
	s.app.Get("/debug/pprof/", adaptor.HTTPHandlerFunc(pprof.Index))
	s.app.Get("/debug/pprof/heap", adaptor.HTTPHandler(pprof.Handler("heap")))
	s.app.Get("/debug/pprof/goroutine", adaptor.HTTPHandler(pprof.Handler("goroutine")))
	s.app.Get("/debug/pprof/allocs", adaptor.HTTPHandler(pprof.Handler("allocs")))

	// 404 handler
	s.app.Use(func(c *fiber.Ctx) error {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": helpers.Problem404(), "data": nil})
	})

	// Run http server
	go func() {
		err := s.app.Listen(cfg.IPService.APIServer.Addr)
		if err != nil {
			s.logger.New("http").Error(err, "listen_error")
		}
	}()

	s.logger.Info("started")

	return s, nil
}

// client: Accept -> server
// server: content-type -> client
func (s *Service) getAccept(c *fiber.Ctx) string {
	accept := c.Get("Accept")
	switch {
	case strings.Contains(accept, MIMEJSON):
		return MIMEJSON
	case strings.Contains(accept, MIMEPlain):
		return MIMEPlain
	case strings.Contains(accept, MIMEHTML):
		return MIMEHTML
	case accept == "*/*" || accept == "":
		return MIMEPlain
	default:
		return MIMEHTML
	}
}

// clientIP extracts the real client IP from the request.
// If proxy_header is configured, it reads the IP from that header.
// Otherwise, it uses the direct remote address.
func (s *Service) clientIP(c *fiber.Ctx) string {
	if s.config.IPService.APIServer.BehindProxy {
		if ip := c.Get("X-Forwarded-For"); ip != "" {
			return ip
		}
	}
	return c.Context().RemoteIP().String()
}

func (s *Service) regEndpoint(ctx context.Context, method, path string, handler func(context.Context, *fiber.Ctx) (any, error)) {
	s.app.Add(method, path, func(c *fiber.Ctx) error {
		clientIP := s.clientIP(c)
		s.logger.Debug("register endpoint", "method", method, "path", path, "clientip", clientIP)
		ctx = contexthandler.Add(ctx, "request", &contexthandler.RequestContext{
			ClientIP:  clientIP,
			UserAgent: string(c.Request().Header.UserAgent()),
			Accept:    s.getAccept(c),
		})
		res, err := handler(ctx, c)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"data": nil, "error": helpers.NewErrorFromError(err)})
		}

		requestValues, err := contexthandler.Get(ctx, "request")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"data": nil, "error": helpers.NewErrorFromError(err)})
		}

		switch requestValues.Accept {
		case MIMEHTML:
			switch r := res.(type) {
			case *model.ReplyIPInformation:
				return c.Render("index", r)
			}
		case MIMEJSON:
			return c.JSON(res)
		case MIMEPlain:
			return c.SendString(fmt.Sprintf("%v\n", res))
		}
		return nil
	})
}

// Close closing httpserver
func (s *Service) Close(ctx context.Context) error {
	s.logger.Info("Quit")
	return s.app.Shutdown()
}
