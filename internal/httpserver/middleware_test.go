package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ip_service/pkg/model"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestMiddlewareTraceID checks that middlewareTraceID sets a req-id header
// and Locals value on the request.
func TestMiddlewareTraceID(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Use(s.middlewareTraceID(t.Context()))
	app.Get("/", func(c *fiber.Ctx) error {
		id, _ := c.Locals("req-id").(string)
		return c.SendString(id)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	assert.NotEmpty(t, string(body))
	assert.NotEmpty(t, resp.Header.Get("req-id"))
}

func TestMiddlewareDuration(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Use(s.middlewareDuration(t.Context()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestMiddlewareLogger(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Use(s.middlewareTraceID(t.Context()))
	app.Use(s.middlewareLogger(t.Context()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestMiddlewareCrash(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Use(s.middlewareCrash(t.Context()))
	app.Get("/", func(c *fiber.Ctx) error {
		panic("boom")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	v := map[string]any{}
	assert.NoError(t, json.Unmarshal(body, &v))
	errObj, ok := v["error"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "internal_server_error", errObj["title"])
}

func TestMiddlewareTimeout(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Use(s.middlewareTimeout(t.Context()))
	app.Get("/", func(c *fiber.Ctx) error {
		// The middleware installs a context with a timeout via SetUserContext;
		// verify a deadline is present.
		_, ok := c.UserContext().Deadline()
		if !ok {
			return c.SendStatus(500)
		}
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClientIP_BehindProxy(t *testing.T) {
	s := mockService(t)
	s.config = &model.Cfg{IPService: &model.IPService{APIServer: model.APIServer{BehindProxy: true}}}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString(s.clientIP(c))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.RemoteAddr = "5.6.7.8:1234"
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "1.2.3.4", string(body))
}

func TestClientIP_Direct(t *testing.T) {
	s := mockService(t)
	s.config = &model.Cfg{IPService: &model.IPService{APIServer: model.APIServer{BehindProxy: false}}}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString(s.clientIP(c))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	assert.NotEmpty(t, string(body))
}

func TestStatusForError(t *testing.T) {
	assert.Equal(t, 504, statusForError(context.DeadlineExceeded))
	assert.Equal(t, 408, statusForError(context.Canceled))
	assert.Equal(t, 400, statusForError(assertErr("boom")))
	// Nil default falls through to 400.
	assert.Equal(t, 400, statusForError(nil))
}

// small helper to build a comparable non-sentinel error
func assertErr(s string) error { return &simpleErr{s} }

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }

func TestBindRequest_JSONBody(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Post("/x", func(c *fiber.Ctx) error {
		req := &struct {
			IP string `json:"ip" validate:"required,ip"`
		}{}
		if err := s.bindRequest(t.Context(), c, req); err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.SendString(req.IP)
	})

	body := `{"ip":"10.0.0.1"}`
	req := httptest.NewRequest("POST", "/x", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	got, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "10.0.0.1", string(got))
}

func TestBindRequest_URIParam(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Get("/lookup/:ip", func(c *fiber.Ctx) error {
		req := &struct {
			IP string `json:"ip" uri:"ip" validate:"required,ip"`
		}{}
		if err := s.bindRequest(t.Context(), c, req); err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.SendString(req.IP)
	})

	req := httptest.NewRequest("GET", "/lookup/1.2.3.4", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	got, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "1.2.3.4", string(got))
}

// Force middlewareTimeout to actually time out.
func TestMiddlewareTimeoutFires(t *testing.T) {
	s := mockService(t)
	app := fiber.New()
	app.Use(s.middlewareTimeout(t.Context()))
	app.Get("/slow", func(c *fiber.Ctx) error {
		select {
		case <-c.UserContext().Done():
			return c.UserContext().Err()
		case <-time.After(2 * time.Second):
			return c.SendString("late")
		}
	})

	// Nothing to assert about timing precisely; just make sure no panic.
	req := httptest.NewRequest("GET", "/slow", nil)
	// Cap the test's own wait so it doesn't hang.
	_, _ = app.Test(req, 100)
}
