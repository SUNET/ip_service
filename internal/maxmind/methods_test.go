package maxmind

import (
	"context"
	"fmt"
	"ip_service/internal/store"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/trace"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	latestTS = "Thu, 01 Sep 2022 18:54:52 GMT"
	oldTS    = "Thu, 01 Aug 2022 18:54:52 GMT"
)

func mockConfig(t *testing.T, storePath string) *model.Cfg {
	cfg := &model.Cfg{
		IPService: &model.IPService{
			APIServer: model.APIServer{
				Addr: "",
			},
			Production: false,
			HTTPProxy:  "",
			Log:        model.Log{},
			MaxMind: model.MaxMind{
				AutomaticUpdate:   true,
				UpdatePeriodicity: 100,
				LicenseKey:        "",
				Enterprise:        false,
				RetryCounter:      0,
				DB: map[string]model.MaxMindDB{
					"asn": {
						FilePath: filepath.Join("..", "..", "testdata", "GeoLite2-asn-Test.mmdb"),
					},
					"city": {
						FilePath: filepath.Join("..", "..", "testdata", "GeoLite2-city-Test.mmdb"),
					},
				},
			},
			Store: model.Store{
				File: model.FileStorage{
					Path: storePath,
				},
			},
			Tracing: model.Tracing{},
		},
	}

	return cfg

}

func mockService(t *testing.T, storeDir string, preOpenDB bool) *Service {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	url := strings.Join([]string{server.URL, "?license_key=%s"}, "/")

	mockGenericEndpointServer(t, mux, "HEAD", fmt.Sprintf(url, "testKey"), nil, 200)

	tracer, err := trace.NoTracing(context.TODO())
	assert.NoError(t, err)

	cfg := mockConfig(t, storeDir)

	store, err := store.New(context.TODO(), cfg, tracer, logger.NewSimple("test-store"))
	assert.NoError(t, err)

	s, err := New(context.TODO(), cfg, store, tracer, logger.NewSimple("test-maxmind"))
	assert.NoError(t, err)

	return s
}

func mockGenericEndpointServer(t *testing.T, mux *http.ServeMux, method, url string, reply []byte, statusCode int) {
	mux.HandleFunc(url,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("last-modified", latestTS)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(statusCode)
			testMethod(t, r, method)
			testURL(t, r, url)
			_, err := w.Write(reply)
			assert.NoError(t, err)
		},
	)
}

func testMethod(t *testing.T, r *http.Request, want string) {
	assert.Equal(t, want, r.Method)
}

func testURL(t *testing.T, r *http.Request, want string) {
	assert.Equal(t, want, r.RequestURI)
}

func TestOpenDB(t *testing.T) {
	type have struct {
		dbType string
	}
	tts := []struct {
		name string
		have have
		want string
	}{
		{
			name: "1-ASN-OK",
			have: have{
				dbType: "asn",
			},
			want: "",
		},
		{
			name: "2-City-OK",
			have: have{
				dbType: "city",
			},
			want: "",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			service := mockService(t, tempDir, false)

			err := service.loadDB(context.TODO(), tt.have.dbType)
			assert.NoError(t, err)
		})
	}
}
