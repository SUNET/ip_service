package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"ip_service/internal/apiv1"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/internal/whois"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/model"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/gofiber/fiber/v2"
	"github.com/google/go-cmp/cmp"
	ua "github.com/mileusna/useragent"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
)

const (
	mockIP           = "89.160.20.112"
	mockUserAgent    = "test-application/3.14.15"
	mockAcceptHeader = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
)

var (
	mockIPWithPort         = fmt.Sprintf("%s:8000", mockIP)
	mockReplyIPInformation = &model.ReplyIPInformation{
		IP:              mockIP,
		IPDecimal:       "1503663216",
		ASN:             29518,
		ASNOrganization: "Bredband2 AB",
		City:            "Linköping",
		Country:         "Sweden",
		CountryISO:      "SE",
		IsEU:            true,
		Region:          "",
		RegionCode:      "",
		PostalCode:      "",
		Coordinates: &model.Coordinates{
			Latitude:  58.4167,
			Longitude: 15.6167,
		},
		Continent: "Europe",
		Timezone:  "Europe/Stockholm",
		Hostname:  "",
		UserAgent: ua.UserAgent{
			Name:      "test-application",
			Version:   "3.14.15",
			OS:        "",
			OSVersion: "",
			Device:    "",
			Mobile:    false,
			Tablet:    false,
			Desktop:   false,
			Bot:       false,
			URL:       "",
			String:    "",
		},
	}
)

func mockHTMLTemplate(t *testing.T) string {
	tmpl, err := template.ParseFiles("../../templates/index.html")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, mockReplyIPInformation)
	assert.NoError(t, err)
	return buf.String()
}

func mockService(t *testing.T) *Service {
	ctx := context.TODO()
	dbCity, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-city-Test.mmdb"))
	assert.NoError(t, err)

	dbASN, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-asn-Test.mmdb"))
	assert.NoError(t, err)

	tracer, err := trace.NewForTesting(ctx, "test", logger.NewSimple("test"))
	assert.NoError(t, err)

	maxmind := &maxmind.Service{
		DBCity: dbCity,
		DBASN:  dbASN,
		TP:     tracer,
		Log:    logger.NewSimple("test-maxmind"),
		DBMeta: map[string]*maxmind.DBObject{
			model.MaxmindDBTypeASN: {
				MU: sync.RWMutex{},
			},
			model.MaxmindDBTypeCity: {
				MU: sync.RWMutex{},
			},
		},
	}
	whois := &whois.Service{}

	store := &store.Service{}

	cfg := &model.Cfg{}

	apiv1, err := apiv1.New(ctx, maxmind, whois, store, cfg, tracer, logger.NewSimple("test-api"))
	assert.NoError(t, err)

	s := &Service{
		config:  &model.Cfg{},
		logger:  logger.NewSimple("test-httpserver"),
		TP:      tracer,
		metrics: &metrics{},
		apiv1:   apiv1,
	}

	s.metrics.init()

	return s
}

func mockContextRequest(accept string) context.Context {
	return contexthandler.Add(context.TODO(), "request", &contexthandler.RequestContext{
		ClientIP:  mockIP,
		UserAgent: mockUserAgent,
		Accept:    accept,
	})
}

func (s *Service) mockEndpointIndexPlain(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEPlain)

	reply, err := s.endpointIndex(ctx, c)
	if err != nil {
		s.logger.Error(err, "mockEndpointIndexPlain")
	}
	s.logger.Debug("mockEndpointIndexPlain", "reply", reply)
	return c.SendString(fmt.Sprintf("%s", reply))
}

func (s *Service) mockEndpointIndexJSON(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEJSON)

	reply, err := s.endpointIndex(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.JSON(reply)
}

func (s *Service) mockEndpointIndexHTML(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEHTML)

	reply, err := s.endpointIndex(ctx, c)
	if err != nil {
		fmt.Println("error:", err)
	}
	return c.Render("index", reply.(*model.ReplyIPInformation))
}

func (s *Service) mockEndpointCityPlain(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEPlain)

	reply, err := s.endpointCity(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.SendString(fmt.Sprintf("%s", reply))
}

func (s *Service) mockEndpointCityJSON(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEJSON)

	reply, err := s.endpointCity(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.JSON(reply)
}

func (s *Service) mockEndpointCountryPlain(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEPlain)

	reply, err := s.endpointCountry(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.SendString(fmt.Sprintf("%s", reply))
}

func (s *Service) mockEndpointCountryJSON(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEJSON)

	reply, err := s.endpointCountry(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.JSON(reply)
}

func (s *Service) mockEndpointCountryISOPlain(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEPlain)

	reply, err := s.endpointCountryISO(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.SendString(fmt.Sprintf("%s", reply))
}

func (s *Service) mockEndpointCountryISOJSON(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEJSON)

	reply, err := s.endpointCountryISO(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.JSON(reply)
}

func (s *Service) mockEndpointASNPlain(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEPlain)

	reply, err := s.endpointASN(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.SendString(fmt.Sprintf("%v", reply))
}

func (s *Service) mockEndpointASNJSON(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEJSON)

	reply, err := s.endpointASN(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.JSON(reply)
}

func (s *Service) mockEndpointCoordinatesPlain(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEPlain)

	reply, err := s.endpointCoordinates(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.SendString(fmt.Sprintf("%v", reply))
}

func (s *Service) mockEndpointCoordinatesJSON(c *fiber.Ctx) error {
	ctx := mockContextRequest(MIMEJSON)

	reply, err := s.endpointCoordinates(ctx, c)
	if err != nil {
		panic(err)
	}
	return c.JSON(reply)
}

func TestEndpoints(t *testing.T) {
	service := mockService(t)
	type have struct {
		path           string
		requestContext *contexthandler.RequestContext
		httpHandler    fiber.Handler
	}
	tts := []struct {
		name string
		have have
		want any
	}{
		{
			name: "index/plain",
			have: have{
				path: "/",
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				httpHandler: service.mockEndpointIndexPlain,
			},
			want: mockIP,
		},
		{
			name: "index/json",
			have: have{
				path: "/",
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				httpHandler: service.mockEndpointIndexJSON,
			},
			want: map[string]any{
				"ip": mockIP,
			},
		},
		{
			name: "index/html",
			have: have{
				path: "/",
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				httpHandler: service.mockEndpointIndexHTML,
			},
			want: mockHTMLTemplate(t),
		},
		{
			name: "city/plain",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/city",
				httpHandler: service.mockEndpointCityPlain,
			},
			want: "Linköping",
		},
		{
			name: "city/json",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/city",
				httpHandler: service.mockEndpointCityJSON,
			},
			want: map[string]any{
				"city": "Linköping",
			},
		},
		{
			name: "city/plain",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/country",
				httpHandler: service.mockEndpointCountryPlain,
			},
			want: "Sweden",
		},
		{
			name: "country/json",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/country",
				httpHandler: service.mockEndpointCountryJSON,
			},
			want: map[string]any{
				"country": "Sweden",
			},
		},
		{
			name: "countryISO/plain",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/country-iso",
				httpHandler: service.mockEndpointCountryISOPlain,
			},
			want: "SE",
		},
		{
			name: "countryISO/json",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/country-iso",
				httpHandler: service.mockEndpointCountryISOJSON,
			},
			want: map[string]any{
				"country_iso": "SE",
			},
		},
		{
			name: "ASN/plain",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/asn",
				httpHandler: service.mockEndpointASNPlain,
			},
			want: fmt.Sprintf("%v", "29518"),
		},
		{
			name: "ASN/json",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/asn",
				httpHandler: service.mockEndpointASNJSON,
			},
			want: map[string]any{
				"asn": float64(29518),
			},
		},
		{
			name: "coordinate/plain",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/coordinates",
				httpHandler: service.mockEndpointCoordinatesPlain,
			},
			want: fmt.Sprintf("%v", []float64{58.4167, 15.6167}),
		},
		{
			name: "coordinate/json",
			have: have{
				requestContext: &contexthandler.RequestContext{
					ClientIP:  mockIP,
					UserAgent: mockUserAgent,
					Accept:    mockAcceptHeader,
				},
				path:        "/coordinates",
				httpHandler: service.mockEndpointCoordinatesJSON,
			},
			want: map[string]any{
				"coordinates": map[string]any{
					"lat":  float64(58.4167),
					"long": float64(15.6167),
				},
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{
				Views:                 newTestViewEngine(t),
				DisableStartupMessage: true,
			})

			app.Get(tt.have.path, tt.have.httpHandler)

			req := httptest.NewRequest("GET", tt.have.path, nil)
			req.Header.Set("User-Agent", tt.have.requestContext.UserAgent)
			req.Header.Set("Accept", tt.have.requestContext.Accept)
			req.RemoteAddr = mockIPWithPort

			resp, err := app.Test(req, -1)
			assert.NoError(t, err)

			got, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			switch tt.have.requestContext.Accept {
			case MIMEJSON:
				v := map[string]any{}
				err := json.Unmarshal(got, &v)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, v)

			case MIMEPlain:
				t.Logf("got:  %q", string(got))
				assert.Equal(t, tt.want, string(got))

			case "*/*", MIMEHTML:
				if diff := cmp.Diff(tt.want, string(got)); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// newTestViewEngine creates an html template engine for tests
func newTestViewEngine(t *testing.T) *testViewEngine {
	t.Helper()
	return &testViewEngine{}
}

type testViewEngine struct{}

func (e *testViewEngine) Load() error { return nil }

func (e *testViewEngine) Render(w io.Writer, name string, data interface{}, layout ...string) error {
	tmpl, err := template.ParseFiles("../../templates/" + name + ".html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}
