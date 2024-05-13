package httpserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
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
			accept:      gin.MIMEJSON,
			contentType: gin.MIMEJSON,
		},
		{
			name:        "2-plain",
			accept:      gin.MIMEPlain,
			contentType: gin.MIMEPlain,
		},
		{
			name:        "3-html",
			accept:      gin.MIMEHTML,
			contentType: gin.MIMEHTML,
		},
		{
			name:        "5-image-*",
			accept:      "image/*",
			contentType: gin.MIMEHTML,
		},
		{
			name:        "6-*-*",
			accept:      "*/*",
			contentType: gin.MIMEPlain,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			service := mockService(t)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
				URL:    &url.URL{},
			}
			c.Request.Header.Set("Accept", tt.accept)

			got := service.getAccept(c)
			assert.Equal(t, tt.contentType, got)
		})
	}
}
