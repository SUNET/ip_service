package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"ip_service/internal/apiv1"
	"ip_service/internal/maxmind"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	ua "github.com/mileusna/useragent"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
)

const (
	mockIP           = "89.160.20.112"
	mockISP          = "89.160.0.0"
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
	dbCity, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-City-Test.mmdb"))
	assert.NoError(t, err)

	dbASN, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-ASN-Test.mmdb"))
	assert.NoError(t, err)
	maxmind := &maxmind.Service{
		DBCity: dbCity,
		DBASN:  dbASN,
	}
	apiv1, err := apiv1.New(context.TODO(), maxmind, nil, nil, logger.New("test-api", false))
	assert.NoError(t, err)

	s := &Service{
		config: &model.Cfg{},
		logger: &logger.Logger{},
		server: &http.Server{},
		apiv1:  apiv1,
		gin:    &gin.Engine{},
	}

	return s
}

func TestNegotiateContentType(t *testing.T) {
	tts := []struct {
		name string
		have string
		want string
	}{
		{
			name: "webbrowser",
			have: mockAcceptHeader,
			want: gin.MIMEHTML,
		},
		{
			name: "empty",
			have: "",
			want: "*/*",
		},
		{
			name: "plain",
			have: gin.MIMEPlain,
			want: gin.MIMEPlain,
		},
		{
			name: "json",
			have: gin.MIMEJSON,
			want: gin.MIMEJSON,
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
			c.Request.Header.Set("Accept", tt.have)

			got := service.negotiateContentType(context.TODO(), c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func (s *Service) mockEndpointIndexPlain(c *gin.Context) {
	reply, err := s.endpointIndex(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.Writer.WriteString(fmt.Sprintf("%v", reply))
}

func (s *Service) mockEndpointIndexNewline(c *gin.Context) {
	reply, err := s.endpointIndex(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.Writer.WriteString(fmt.Sprintf("%v\n", reply))
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
	c.Writer.WriteString(fmt.Sprintf("%s", reply))
}

func (s *Service) mockEndpointCityNewline(c *gin.Context) {
	reply, err := s.endpointCity(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.Writer.WriteString(fmt.Sprintf("%v\n", reply))
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
	c.Writer.WriteString(fmt.Sprintf("%s", reply))
}

func (s *Service) mockEndpointCountryNewline(c *gin.Context) {
	reply, err := s.endpointCountry(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.Writer.WriteString(fmt.Sprintf("%v\n", reply))
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
	c.Writer.WriteString(fmt.Sprintf("%s", reply))
}

func (s *Service) mockEndpointCountryISONewline(c *gin.Context) {
	reply, err := s.endpointCountryISO(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.Writer.WriteString(fmt.Sprintf("%v\n", reply))
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
	c.Writer.WriteString(fmt.Sprintf("%v", reply))
}

func (s *Service) mockEndpointASNNewline(c *gin.Context) {
	reply, err := s.endpointASN(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.Writer.WriteString(fmt.Sprintf("%v\n", reply))
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
	c.Writer.WriteString(fmt.Sprintf("%v", reply))
}

func (s *Service) mockEndpointCoordinatesNewline(c *gin.Context) {
	reply, err := s.endpointCoordinates(context.TODO(), c)
	if err != nil {
		panic(err)
	}
	c.Writer.WriteString(fmt.Sprintf("%v\n", reply))
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
		ip          string
		path        string
		mimeType    string
		userAgent   string
		httpHandler func(c *gin.Context)
	}
	tts := []struct {
		have have
		want interface{}
	}{
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/",
				mimeType:    gin.MIMEPlain,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointIndexPlain,
			},
			want: mockIP,
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/",
				mimeType:    "*/*",
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointIndexNewline,
			},
			want: fmt.Sprintf("%v\n", mockIP),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/",
				mimeType:    gin.MIMEJSON,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointIndexJSON,
			},
			want: map[string]interface{}{
				"ip": mockIP,
			},
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/",
				mimeType:    gin.MIMEHTML,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointIndexHTML,
			},
			want: mockHTMLTemplate(t),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/city",
				mimeType:    gin.MIMEPlain,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCityPlain,
			},
			want: "Linköping",
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/city",
				mimeType:    "*/*",
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCityNewline,
			},
			want: fmt.Sprintf("%v\n", "Linköping"),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/city",
				mimeType:    gin.MIMEJSON,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCityJSON,
			},
			want: map[string]interface{}{
				"city": "Linköping",
			},
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/country",
				mimeType:    gin.MIMEPlain,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCountryPlain,
			},
			want: "Sweden",
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/country",
				mimeType:    "*/*",
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCountryNewline,
			},
			want: fmt.Sprintf("%v\n", "Sweden"),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/country",
				mimeType:    gin.MIMEJSON,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCountryJSON,
			},
			want: map[string]interface{}{
				"country": "Sweden",
			},
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/country-iso",
				mimeType:    gin.MIMEPlain,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCountryISOPlain,
			},
			want: "SE",
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/country-iso",
				mimeType:    "*/*",
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCountryISONewline,
			},
			want: fmt.Sprintf("%v\n", "SE"),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/country-iso",
				mimeType:    gin.MIMEJSON,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCountryISOJSON,
			},
			want: map[string]interface{}{
				"country_iso": "SE",
			},
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/asn",
				mimeType:    gin.MIMEPlain,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointASNPlain,
			},
			want: fmt.Sprintf("%v", "29518"),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/asn",
				mimeType:    "*/*",
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointASNNewline,
			},
			want: fmt.Sprintf("%v\n", "29518"),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/asn",
				mimeType:    gin.MIMEJSON,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointASNJSON,
			},
			want: map[string]interface{}{
				"asn": float64(29518),
			},
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/coordinates",
				mimeType:    gin.MIMEPlain,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCoordinatesPlain,
			},
			want: fmt.Sprintf("%v",[]float64{58.4167, 15.6167}),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/coordinates",
				mimeType:    "*/*",
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCoordinatesNewline,
			},
			want: fmt.Sprintf("%v\n", []float64{58.4167, 15.6167}),
		},
		{
			have: have{
				ip:          mockIPWithPort,
				path:        "/coordinates",
				mimeType:    gin.MIMEJSON,
				userAgent:   mockUserAgent,
				httpHandler: service.mockEndpointCoordinatesJSON,
			},
			want: map[string]interface{}{
				"coordinates": map[string]interface{}{
					"lat":  float64(58.4167),
					"long": float64(15.6167),
				},
			},
		},
	}

	for _, tt := range tts {
		t.Run(fmt.Sprintf("%q --> %q", tt.have.path, tt.have.mimeType), func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			r := gin.Default()
			r.LoadHTMLGlob("../../templates/*.html")
			r.Static("/assets", "./assets")

			r.GET(tt.have.path, tt.have.httpHandler)

			req, err := http.NewRequest("GET", tt.have.path, nil)
			assert.NoError(t, err)

			req.Header.Set("user-agent", tt.have.userAgent)
			req.Header.Set("Accept", tt.have.mimeType)
			req.RemoteAddr = tt.have.ip

			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			got, err := ioutil.ReadAll(w.Body)
			assert.NoError(t, err)

			switch tt.have.mimeType {
			case gin.MIMEJSON:
				v := map[string]interface{}{}
				err := json.Unmarshal(got, &v)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, v)

			case gin.MIMEPlain:
				assert.Equal(t, tt.want, string(got))

			case "*/*":
				if diff := cmp.Diff(tt.want, string(got)); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}

			case gin.MIMEHTML:
				if diff := cmp.Diff(tt.want, string(got)); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
