package httpserver

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAccept(t *testing.T) {
	tts := []struct {
		name        string
		accept      string
		contentType string
	}{
		{
			name:        "1-json",
			accept:      MIMEJSON,
			contentType: MIMEJSON,
		},
		{
			name:        "2-plain",
			accept:      MIMEPlain,
			contentType: MIMEPlain,
		},
		{
			name:        "3-html",
			accept:      MIMEHTML,
			contentType: MIMEHTML,
		},
		{
			name:        "5-image-*",
			accept:      "image/*",
			contentType: MIMEHTML,
		},
		{
			name:        "6-*-*",
			accept:      "*/*",
			contentType: MIMEPlain,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			service := mockService(t)

			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				got := service.getAccept(c)
				return c.SendString(got)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", tt.accept)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)

			buf := make([]byte, 256)
			n, _ := resp.Body.Read(buf)
			got := string(buf[:n])
			assert.Equal(t, tt.contentType, got)
		})
	}
}
