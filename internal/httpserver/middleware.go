package httpserver

import (
	"context"
	"ip_service/pkg/helpers"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"
)

func (s *Service) middlewareDuration(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		duration := time.Since(t)
		c.Set("duration", duration)
	}
}

func (s *Service) middlewareTraceID(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("req-id", shortuuid.New())
		c.Header("req-id", c.GetString("req-id"))
		c.Next()
	}
}

func (s *Service) middlewareLogger(ctx context.Context) gin.HandlerFunc {
	log := s.logger.New("http")
	return func(c *gin.Context) {
		c.Next()
		log.Info("request", "status", c.Writer.Status(), "url", c.Request.URL.String(), "method", c.Request.Method, "req-id", c.GetString("req-id"))
	}
}

func (s *Service) middlewareCrash(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				status := c.Writer.Status()
				s.logger.Debug("crash", "status", status)
				c.JSON(500, gin.H{"data": nil, "error": helpers.NewError("internal_server_error")})
			}
		}()
		c.Next()
	}
}
