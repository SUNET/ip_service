package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"ip_service/internal/apiv1"
	"ip_service/internal/lctree"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/internal/whois"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/model"
	"ip_service/pkg/rpsl"

	"path/filepath"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/gofiber/fiber/v2"
	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// mockServiceWithWhois builds a Service with a working whois tree so
// endpointWhois / endpointLookUpIP work end-to-end.
func mockServiceWithWhois(t *testing.T) *Service {
	t.Helper()

	// Use a private prometheus registry per-test to avoid duplicate registration.
	reg := prometheus.NewRegistry()
	prev := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	t.Cleanup(func() { prometheus.DefaultRegisterer = prev })

	ctx := t.Context()
	dbCity, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-city-Test.mmdb"))
	assert.NoError(t, err)
	dbASN, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-asn-Test.mmdb"))
	assert.NoError(t, err)

	tp, err := trace.NewForTesting(ctx, "t", logger.NewSimple("t"))
	assert.NoError(t, err)

	// Storage-backed store for Status endpoint.
	stCfg := &model.Cfg{IPService: &model.IPService{Store: model.Store{File: model.FileStorage{Path: t.TempDir()}}}}
	st, err := store.New(ctx, stCfg, tp, logger.NewSimple("store"))
	assert.NoError(t, err)

	mm := &maxmind.Service{
		DBCity: dbCity,
		DBASN:  dbASN,
		TP:     tp,
		Log:    logger.NewSimple("mm"),
		DBMeta: map[string]*maxmind.DBObject{
			model.MaxmindDBTypeASN:  {MU: sync.RWMutex{}},
			model.MaxmindDBTypeCity: {MU: sync.RWMutex{}},
		},
	}

	rc := rpsl.RouterClass{
		"89.160.0.0/16": rpsl.ASN{"AS29518": &rpsl.Object{Network: "89.160.0.0/16", Origin: "AS29518"}},
	}
	tree := lctree.New(logger.NewSimple("lctree"))
	assert.NoError(t, tree.Build(ctx, rc))
	ws := whois.NewTestService(tree, rc)

	api, err := apiv1.New(ctx, mm, ws, st, &model.Cfg{}, tp, logger.NewSimple("apiv1"))
	assert.NoError(t, err)

	s := &Service{
		config:  &model.Cfg{IPService: &model.IPService{APIServer: model.APIServer{}}},
		logger:  logger.NewSimple("http"),
		TP:      tp,
		metrics: &metrics{},
		apiv1:   api,
	}
	s.metrics.init()
	return s
}

// registerHandler is a helper wrapping an endpoint with a mock request context.
func registerHandler(app *fiber.App, path string, s *Service, accept string, h func(context.Context, *fiber.Ctx) (any, error)) {
	app.Get(path, func(c *fiber.Ctx) error {
		ctx := contexthandler.Add(c.UserContext(), "request", &contexthandler.RequestContext{
			ClientIP:  "89.160.20.112",
			UserAgent: "ua/1.0",
			Accept:    accept,
		})
		res, err := h(ctx, c)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		if accept == MIMEJSON {
			return c.JSON(res)
		}
		return c.SendString(fmt.Sprintf("%v", res))
	})
}

func TestEndpointAll_JSON(t *testing.T) {
	s := mockServiceWithWhois(t)
	app := fiber.New()
	registerHandler(app, "/all", s, MIMEJSON, s.endpointAll)

	req := httptest.NewRequest("GET", "/all", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	v := map[string]any{}
	assert.NoError(t, json.Unmarshal(body, &v))
	assert.Equal(t, "89.160.20.112", v["ip"])
}

func TestEndpointHealth_JSON(t *testing.T) {
	s := mockServiceWithWhois(t)
	app := fiber.New()
	registerHandler(app, "/health", s, MIMEJSON, s.endpointHealth)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	v := map[string]any{}
	assert.NoError(t, json.Unmarshal(body, &v))
	assert.Contains(t, v, "data")
}

func TestEndpointCollision(t *testing.T) {
	s := mockServiceWithWhois(t)
	app := fiber.New()

	app.Post("/collision", func(c *fiber.Ctx) error {
		ctx := contexthandler.Add(c.UserContext(), "request", &contexthandler.RequestContext{
			ClientIP: "89.160.20.112",
			Accept:   MIMEJSON,
		})
		res, err := s.endpointCollision(ctx, c)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.JSON(res)
	})

	body := `{"ip_1":"10.0.0.0/8","ip_2":"10.1.0.0/16"}`
	req := httptest.NewRequest("POST", "/collision", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	got := map[string]any{}
	buf, _ := io.ReadAll(resp.Body)
	assert.NoError(t, json.Unmarshal(buf, &got))
	assert.Equal(t, true, got["collision"])
}

func TestEndpointCollision_BadInput(t *testing.T) {
	s := mockServiceWithWhois(t)
	app := fiber.New()
	app.Post("/collision", func(c *fiber.Ctx) error {
		ctx := contexthandler.Add(c.UserContext(), "request", &contexthandler.RequestContext{
			ClientIP: "89.160.20.112",
			Accept:   MIMEJSON,
		})
		_, err := s.endpointCollision(ctx, c)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.SendStatus(200)
	})

	body := `{"ip_1":"notcidr","ip_2":"10.1.0.0/16"}`
	req := httptest.NewRequest("POST", "/collision", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestEndpointWhois(t *testing.T) {
	s := mockServiceWithWhois(t)
	app := fiber.New()

	app.Get("/whois/:ip", func(c *fiber.Ctx) error {
		ctx := contexthandler.Add(c.UserContext(), "request", &contexthandler.RequestContext{
			ClientIP: "89.160.20.112",
			Accept:   MIMEJSON,
		})
		res, err := s.endpointWhois(ctx, c)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.JSON(res)
	})

	req := httptest.NewRequest("GET", "/whois/89.160.20.112", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	buf, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(buf), "AS29518")
}

func TestEndpointLookUpIP(t *testing.T) {
	s := mockServiceWithWhois(t)
	app := fiber.New()

	app.Get("/lookup/:ip", func(c *fiber.Ctx) error {
		ctx := contexthandler.Add(c.UserContext(), "request", &contexthandler.RequestContext{
			ClientIP: "127.0.0.1",
			Accept:   MIMEJSON,
		})
		res, err := s.endpointLookUpIP(ctx, c)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.JSON(res)
	})

	req := httptest.NewRequest("GET", "/lookup/89.160.20.112", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	buf, _ := io.ReadAll(resp.Body)
	v := map[string]any{}
	assert.NoError(t, json.Unmarshal(buf, &v))
	assert.Equal(t, "89.160.20.112", v["ip"])
}

func TestClose(t *testing.T) {
	s := mockService(t)
	// Provide a fiber app for Close to shut down.
	s.app = fiber.New()
	assert.NoError(t, s.Close(t.Context()))
}
