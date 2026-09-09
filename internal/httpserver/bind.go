package httpserver

import (
	"context"
	"ip_service/pkg/helpers"
	"mime"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

func (s *Service) bindRequest(ctx context.Context, c *fiber.Ctx, v interface{}) error {
	ctx, span := s.TP.Start(ctx, "httpserver:bindRequest")
	defer span.End()

	// Parse the media type so charset and other params (e.g., "application/json; charset=utf-8") are accepted.
	mediaType, _, _ := mime.ParseMediaType(c.Get("Content-Type"))
	if mediaType == fiber.MIMEApplicationJSON {
		if err := c.BodyParser(v); err != nil {
			return err
		}
	}

	// Bind query parameters
	if err := c.QueryParser(v); err != nil {
		return err
	}

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
