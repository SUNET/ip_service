package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"ip_service/internal/apiv1"
	"ip_service/internal/maxmind"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/trace"

	"github.com/gin-gonic/gin"
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
		Latitude:        58.4167,
		Longitude:       15.6167,
		Timezone:        "Europe/Stockholm",
		Hostname:        "",
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
		Continent: "Europe",
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
	dbCity, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-city-Test.mmdb"))
	assert.NoError(t, err)

	dbASN, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-asn-Test.mmdb"))
	assert.NoError(t, err)

	tracer, err := trace.NoTracing(context.TODO())
	assert.NoError(t, err)

	maxmind := &maxmind.Service{
		DBCity: dbCity,
		DBASN:  dbASN,
		TP:     tracer,
		DBMeta: map[string]*maxmind.DBObject{
			"city": {
				MU: sync.RWMutex{},
			},
			"asn": {
				MU: sync.RWMutex{},
			},
		},
	}

	apiv1, err := apiv1.New(context.TODO(), maxmind, nil, nil, tracer, logger.NewSimple("test-api"))
	assert.NoError(t, err)

	s := &Service{
		config: &model.Cfg{},
		logger: &logger.Log{},
		TP:     tracer,
		server: &http.Server{
			ReadHeaderTimeout: 1 * time.Second,
		},
		apiv1: apiv1,
		gin:   &gin.Engine{},
	}

	return s
}

func (s *Service) mockEndpointIndexPlain(c *gin.Context) {
	reply, err := s.endpointIndex(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.WriteString(fmt.Sprintf("%v", reply))
	if err != nil {
		panic(err)
	}
}

func (s *Service) mockEndpointIndexJSON(c *gin.Context) {
	reply, err := s.endpointIndex(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.JSON(200, reply)
}

func (s *Service) mockEndpointIndexHTML(c *gin.Context) {
	reply, err := s.endpointIndex(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.HTML(http.StatusOK, "index.html", reply.(*model.ReplyIPInformation))
}

func (s *Service) mockEndpointCityPlain(c *gin.Context) {
	reply, err := s.endpointCity(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.WriteString(fmt.Sprintf("%s", reply))
	if err != nil {
		panic(err)
	}
}

func (s *Service) mockEndpointCityJSON(c *gin.Context) {
	reply, err := s.endpointCity(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.JSON(200, reply)
}

func (s *Service) mockEndpointCountryPlain(c *gin.Context) {
	reply, err := s.endpointCountry(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.WriteString(fmt.Sprintf("%s", reply))
	if err != nil {
		panic(err)
	}
}

func (s *Service) mockEndpointCountryJSON(c *gin.Context) {
	reply, err := s.endpointCountry(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.JSON(200, reply)
}

func (s *Service) mockEndpointCountryISOPlain(c *gin.Context) {
	reply, err := s.endpointCountryISO(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.WriteString(fmt.Sprintf("%s", reply))
	if err != nil {
		panic(err)
	}
}

func (s *Service) mockEndpointCountryISOJSON(c *gin.Context) {
	reply, err := s.endpointCountryISO(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.JSON(200, reply)
}

func (s *Service) mockEndpointASNPlain(c *gin.Context) {
	reply, err := s.endpointASN(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.WriteString(fmt.Sprintf("%v", reply))
	if err != nil {
		panic(err)
	}
}

func (s *Service) mockEndpointASNJSON(c *gin.Context) {
	reply, err := s.endpointASN(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.JSON(200, reply)
}

func (s *Service) mockEndpointCoordinatesPlain(c *gin.Context) {
	reply, err := s.endpointCoordinates(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	_, err = c.Writer.WriteString(fmt.Sprintf("%v", reply))
	if err != nil {
		panic(err)
	}
}

func (s *Service) mockEndpointCoordinatesJSON(c *gin.Context) {
	reply, err := s.endpointCoordinates(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.JSON(200, reply)
}

func TestEndpoints(t *testing.T) {
	service := mockService(t)
	type have struct {
		path           string
		requestContext *contexthandler.RequestContext
		httpHandler    func(c *gin.Context)
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
					ClientIP:  mockIPWithPort,
					LookupIP:  "",
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
					ClientIP:  mockIPWithPort,
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
			gin.SetMode(gin.TestMode)

			r := gin.Default()
			r.LoadHTMLGlob("../../templates/*.html")
			r.Static("/assets", "./assets")

			r.GET(tt.have.path, tt.have.httpHandler)

			req, err := http.NewRequest("GET", tt.have.path, nil)
			assert.NoError(t, err)

			req.Header.Set("user-agent", tt.have.requestContext.UserAgent)
			req.Header.Set("Accept", tt.have.requestContext.Accept)
			req.RemoteAddr = tt.have.requestContext.ClientIP

			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			got, err := io.ReadAll(w.Body)
			assert.NoError(t, err)

			switch tt.have.requestContext.Accept {
			case gin.MIMEJSON:
				v := map[string]any{}
				err := json.Unmarshal(got, &v)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, v)

			case gin.MIMEPlain:
				assert.Equal(t, tt.want, string(got))

			case "*/*", gin.MIMEHTML:
				if diff := cmp.Diff(tt.want, string(got)); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
