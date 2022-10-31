package maxmind

import (
	"context"
	"fmt"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
)

const (
	latest_ts = "Thu, 01 Sep 2022 18:54:52 GMT"
	old_ts    = "Thu, 01 Aug 2022 18:54:52 GMT"
)

func mockService(t *testing.T, dir string) *Service {
	asnFilePath := filepath.Join("..", "..", "testdata", "GeoLite2-ASN-Test.mmdb")
	cityFilePath := filepath.Join("..", "..", "testdata", "GeoLite2-City-Test.mmdb")
	dbCity, err := geoip2.Open(cityFilePath)
	assert.NoError(t, err)

	dbASN, err := geoip2.Open(asnFilePath)
	assert.NoError(t, err)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	url := strings.Join([]string{server.URL, "?license_key=%s"}, "/")

	mockGenericEndpointServer(t, mux, "HEAD", fmt.Sprintf(url, "testKey"), nil, 200)

	s := &Service{
		cfg: &model.Cfg{
			MaxMind: struct {
				AutomaticUpdate   bool          "yaml:\"automatic_update\""
				UpdatePeriodicity time.Duration "yaml:\"update_periodicity\""
				LicenseKey        string        "yaml:\"license_key\""
				Enterprise        bool          "yaml:\"enterprise\""
			}{
				UpdatePeriodicity: 10,
				LicenseKey:        "testKey",
				Enterprise:        false,
			},
			Store: struct {
				File struct {
					Path string "yaml:\"path\""
				} "yaml:\"file\""
			}{
				File: struct {
					Path string "yaml:\"path\""
				}{
					Path: dir,
				},
			},
		},
		log: logger.New("test", false).New("test"),
		dbMeta: map[string]dbObject{
			"ASN": {
				url:      url,
				filePath: asnFilePath,
			},
			"City": {
				url:      url,
				filePath: cityFilePath,
			},
		},
		DBCity:   dbCity,
		DBASN:    dbASN,
		kvStore:  nil,
		quitChan: make(chan bool),
	}

	return s
}

func mockGenericEndpointServer(t *testing.T, mux *http.ServeMux, method, url string, reply []byte, statusCode int) {
	mux.HandleFunc(url,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("last-modified", latest_ts)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(statusCode)
			testMethod(t, r, method)
			testURL(t, r, url)
			w.Write(reply)
		},
	)
}

func testMethod(t *testing.T, r *http.Request, want string) {
	assert.Equal(t, want, r.Method)
}

func testURL(t *testing.T, r *http.Request, want string) {
	assert.Equal(t, want, r.RequestURI)
}

func TestGetRemoteVersion(t *testing.T) {
	tts := []struct {
		name string
		have string
		want string
	}{
		{
			name: "OK",
			want: "mura",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			service := mockService(t, tempDir)
			got, err := service.getRemoteVersion(context.TODO(), "ASN")
			fmt.Println("got", got)
			assert.NoError(t, err)
			//assert.Equal(t, tt.want, got)
		})
	}
}

func TestOpenDB(t *testing.T) {
	t.SkipNow()
	type have struct {
		dbType string
	}
	tts := []struct {
		name string
		have have
		want string
	}{
		{
			name: "OK",
			have: have{
				dbType: "ASN",
			},
			want: "",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			service := mockService(t, tempDir)

			err := service.openDB(context.TODO(), tt.have.dbType)
			assert.NoError(t, err)
		})
	}
}
