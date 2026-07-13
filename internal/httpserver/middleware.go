package httpserver

import (
	"context"
	"ip_service/pkg/helpers"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lithammer/shortuuid/v4"
)

const defaultRequestTimeout = 30 * time.Second

func (s *Service) middlewareTimeout(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), defaultRequestTimeout)
		defer cancel()
		c.SetUserContext(ctx)
		return c.Next()
	}
}

func (s *Service) middlewareDuration(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		t := time.Now()
		err := c.Next()
		duration := time.Since(t)
		c.Locals("duration", duration)
		return err
	}
}

func (s *Service) middlewareTraceID(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID := shortuuid.New()
		c.Locals("req-id", reqID)
		c.Set("req-id", reqID)
		return c.Next()
	}
}

func (s *Service) middlewareLogger(ctx context.Context) fiber.Handler {
	log := s.logger.New("http")
	return func(c *fiber.Ctx) error {
		err := c.Next()
		reqID, _ := c.Locals("req-id").(string)
		log.Info("request", "status", c.Response().StatusCode(), "url", c.OriginalURL(), "method", c.Method(), "req-id", reqID)
		return err
	}
}

func (s *Service) middlewareCrash(ctx context.Context) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Debug("crash", "panic", r)
				err = c.Status(500).JSON(fiber.Map{"data": nil, "error": helpers.NewError("internal_server_error")})
			}
		}()
		return c.Next()
	}
}
