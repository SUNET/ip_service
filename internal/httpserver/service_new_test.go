package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"ip_service/internal/apiv1"
	"ip_service/pkg/model"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// TestRegEndpoint_ContentTypes drives regEndpoint for JSON, Plain and HTML.
func TestRegEndpoint_ContentTypes(t *testing.T) {
	s := mockServiceWithWhois(t)
	s.app = fiber.New(fiber.Config{
		Views:                 newTestViewEngine(t),
		DisableStartupMessage: true,
	})
	s.regEndpoint(t.Context(), "GET", "/", s.endpointIndex)

	for _, accept := range []string{MIMEJSON, MIMEPlain, MIMEHTML} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", accept)
		req.RemoteAddr = "89.160.20.112:1234"
		resp, err := s.app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode, "accept=%s", accept)
		body, _ := io.ReadAll(resp.Body)
		assert.NotEmpty(t, body, "accept=%s", accept)
	}
}

func TestNew_ClosesCleanly(t *testing.T) {
	ms := mockServiceWithWhois(t)

	// Swap in a brand-new prometheus registry so New()'s metrics.init doesn't
	// collide with metrics from mockServiceWithWhois.
	reg2 := prometheus.NewRegistry()
	prev := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg2
	t.Cleanup(func() { prometheus.DefaultRegisterer = prev })

	cfg := &model.Cfg{IPService: &model.IPService{
		APIServer:  model.APIServer{Addr: "127.0.0.1:0"},
		Production: true,
	}}

	api := ms.apiv1.(*apiv1.Client)
	svc, err := New(t.Context(), cfg, api, ms.TP, ms.logger)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Trigger the app-level 404 handler via app.Test.
	req := httptest.NewRequest("GET", "/does-not-exist", nil)
	resp, err := svc.app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Static asset lookup for a missing file returns 404.
	req = httptest.NewRequest("GET", "/assets/nope.css", nil)
	resp, err = svc.app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	assert.NoError(t, svc.Close(t.Context()))
}

// Ensure endpointIndex returns a helpful error when Accept is unsupported.
func TestEndpointIndex_UnsupportedAccept(t *testing.T) {
	s := mockServiceWithWhois(t)
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		// Provide a request-context with an unsupported Accept value.
		reqCtx := contextRequestWith(ctx, "boo/hoo")
		_, err := s.endpointIndex(reqCtx, c)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "unsupported")
}

// contextRequestWith attaches a mock request context with a specific Accept value.
func contextRequestWith(ctx context.Context, accept string) context.Context {
	return mockContextRequest(accept)
}

func TestClientIP_UsesRemoteAddrWhenNoXFF(t *testing.T) {
	s := mockService(t)
	s.config = &model.Cfg{IPService: &model.IPService{APIServer: model.APIServer{BehindProxy: true}}}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString(fmt.Sprintf("%s", s.clientIP(c)))
	})

	// BehindProxy is true but no X-Forwarded-For header set → falls back to remote.
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.11.12.13:8080"
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
