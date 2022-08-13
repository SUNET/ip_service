package apiv1

import (
	"context"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"path/filepath"
	"testing"

	ua "github.com/mileusna/useragent"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
)

const (
	mockIP        = "2a02:d040::"
	mockISP       = "89.160.0.0"
	mockUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) QtWebEngine/5.15.2 Chrome/83.0.4103.122 Safari/537.36"
)

func mockClient(t *testing.T) *Client {
	dbCity, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-City-Test.mmdb"))
	assert.NoError(t, err)

	dbASN, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-ASN-Test.mmdb"))
	assert.NoError(t, err)

	c := &Client{
		config: &model.Cfg{},
		logger: &logger.Logger{},
		max: &maxmind.Service{
			DBCity: dbCity,
			DBASN:  dbASN,
		},
		store: &store.Service{},
	}

	return c
}

func TestIPText(t *testing.T) {
	tts := []struct {
		name string
		have string
		want string
	}{
		{
			name: "OK",
			have: "127.0.0.1",
			want: "127.0.0.1",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.IPText(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIPJSON(t *testing.T) {
	tts := []struct {
		name string
		have string
		want map[string]interface{}
	}{
		{
			name: "OK",
			have: "127.0.0.1",
			want: map[string]interface{}{
				"ip": "127.0.0.1",
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.IPJSON(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountryText(t *testing.T) {
	tts := []struct {
		name string
		have string
		want string
	}{
		{
			name: "OK",
			have: mockIP,
			want: "Sweden",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.CountryText(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountryJSON(t *testing.T) {
	tts := []struct {
		name string
		have string
		want map[string]interface{}
	}{
		{
			name: "OK",
			have: mockIP,
			want: map[string]interface{}{
				"country": "Sweden",
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.CountryJSON(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestASNText(t *testing.T) {
	tts := []struct {
		name string
		have string
		want uint
	}{
		{
			name: "OK",
			have: mockISP,
			want: 29518,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.ASNText(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestASNJSON(t *testing.T) {
	tts := []struct {
		name string
		have string
		want map[string]interface{}
	}{
		{
			name: "OK",
			have: mockISP,
			want: map[string]interface{}{
				"asn": uint(29518),
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.ASNJSON(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountryISOText(t *testing.T) {
	tts := []struct {
		name string
		have string
		want string
	}{
		{
			name: "OK",
			have: mockIP,
			want: "SE",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.CountryISOText(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountryISOJSON(t *testing.T) {
	tts := []struct {
		name string
		have string
		want map[string]interface{}
	}{
		{
			name: "OK",
			have: mockIP,
			want: map[string]interface{}{
				"country_iso": "SE",
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.CountryISOJSON(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoordinatesText(t *testing.T) {
	tts := []struct {
		name string
		have string
		want []float64
	}{
		{
			name: "OK",
			have: mockIP,
			want: []float64{62, 15},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.CoordinatesText(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoordinatesJSON(t *testing.T) {
	tts := []struct {
		name string
		have string
		want map[string]interface{}
	}{
		{
			name: "OK",
			have: mockIP,
			want: map[string]interface{}{
				"coordinates": map[string]interface{}{
					"lat":  float64(62),
					"long": float64(15),
				},
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			client := mockClient(t)
			got, err := client.CoordinatesJSON(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllJSON(t *testing.T) {
	tts := []struct {
		name string
		have string
		want *model.ReplyIPInformation
	}{
		{
			name: "OK - IP",
			have: mockIP,
			want: &model.ReplyIPInformation{
				IP:              mockIP,
				IPDecimal:       "55842184228483496777582744539513225216",
				ASN:             0,
				ASNOrganization: "",
				City:            "",
				Country:         "Sweden",
				CountryISO:      "SE",
				IsEU:            true,
				Region:          "",
				RegionCode:      "",
				PostalCode:      "",
				Latitude:        62,
				Longitude:       15,
				Timezone:        "Europe/Stockholm",
				Hostname:        "",
				UserAgent: ua.UserAgent{
					Name:      "QtWebEngine",
					Version:   "5.15.2",
					OS:        "Linux",
					OSVersion: "x86_64",
					Device:    "",
					Mobile:    false,
					Tablet:    false,
					Desktop:   true,
					Bot:       false,
					URL:       "",
					String:    mockUserAgent,
				},
				Continent: "Europe",
			},
		},
		{
			name: "OK - ASN",
			have: mockISP,
			want: &model.ReplyIPInformation{
				IP:              mockISP,
				IPDecimal:       "1503657984",
				ASN:             29518,
				ASNOrganization: "Bredband2 AB",
				UserAgent: ua.UserAgent{
					Name:      "QtWebEngine",
					Version:   "5.15.2",
					OS:        "Linux",
					OSVersion: "x86_64",
					Device:    "",
					Mobile:    false,
					Tablet:    false,
					Desktop:   true,
					Bot:       false,
					URL:       "",
					String:    mockUserAgent,
				},
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), "ip", tt.have)
			ctx = context.WithValue(ctx, "ua", mockUserAgent)

			client := mockClient(t)
			got, err := client.AllJSON(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
