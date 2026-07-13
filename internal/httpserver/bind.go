package httpserver

import (
	"context"
	"ip_service/pkg/helpers"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

func (s *Service) bindRequest(ctx context.Context, c *fiber.Ctx, v interface{}) error {
	ctx, span := s.TP.Start(ctx, "httpserver:bindRequest")
	defer span.End()

	// Bind JSON body if content type is JSON
	if c.Get("Content-Type") == "application/json" {
		_ = c.BodyParser(v)
	}

	// Bind query parameters
	_ = c.QueryParser(v)

	// Bind URI parameters (path params)
	s.bindURIParams(c, v)

	// Validate the bound struct
	if err := helpers.Check(v); err != nil {
		return err
	}

	return nil
}

// bindURIParams maps path parameters to struct fields tagged with `uri:"name"`
func (s *Service) bindURIParams(c *fiber.Ctx, v interface{}) {
	refV := reflect.ValueOf(v).Elem()
	refT := refV.Type()
	for i := 0; i < refT.NumField(); i++ {
		field := refT.Field(i)
		uriTag := field.Tag.Get("uri")
		if uriTag == "" {
			continue
		}
		param := c.Params(uriTag)
		if param != "" {
			refV.Field(i).SetString(param)
		}
	}
}
