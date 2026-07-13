package rpslsource

import (
	"context"
	"ip_service/internal/store"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/model"
	"github.com/SUNET/vc/pkg/trace"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func mockService(t *testing.T, sourceCfg Config) *Service {
	ctx := context.TODO()
	tmpDir := t.TempDir()

	cfg := &model.Cfg{
		IPService: &model.IPService{
			Store: model.Store{
				File: model.FileStorage{
					Path: tmpDir,
				},
			},
			Radb: model.Radb{
				FilePath: tmpDir + "/radb.db",
			},
			RIPE: model.RIPE{
				FilePath: tmpDir + "/ripe.db",
			},
		},
	}

	tp, err := trace.NewForTesting(ctx, "test", logger.NewSimple("test"))
	assert.NoError(t, err)

	log := logger.NewSimple(sourceCfg.Name + "_test")

	st, err := store.New(ctx, cfg, tp, log)
	assert.NoError(t, err)

	sourceCfg.FilePath = tmpDir + "/" + sourceCfg.Name + ".db"

	service, err := New(ctx, st, cfg, log, sourceCfg)
	assert.NoError(t, err)

	// Override rate limit for tests
	service.rateLimit = *rate.NewLimiter(rate.Every(1*time.Second), 5)

	return service
}

func TestGetRemoteSerialHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/CURRENTSERIAL":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("12345678"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	service := mockService(t, Config{
		Name:      "ripe",
		Transport: TransportHTTP,
		RemoteFiles: []RemoteFile{
			{Name: "route", Path: "/route.gz"},
			{Name: "route6", Path: "/route6.gz"},
		},
		SerialPath:   "/CURRENTSERIAL",
		Host:         ts.URL,
		AddEOFMarker: true,
	})

	serial, err := service.getRemoteSerial(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, "12345678", serial)
}

func TestGetRemoteSerialHTTP_RateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	service := mockService(t, Config{
		Name:      "ripe",
		Transport: TransportHTTP,
		RemoteFiles: []RemoteFile{
			{Name: "route", Path: "/route.gz"},
		},
		SerialPath: "/CURRENTSERIAL",
		Host:       ts.URL,
	})
	// Use a plain client to avoid retryablehttp retrying 429s forever
	service.httpClient = &http.Client{Timeout: 5 * time.Second}

	_, err := service.getRemoteSerial(context.TODO())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
}

func TestGetRemoteSerialHTTP_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer ts.Close()

	service := mockService(t, Config{
		Name:      "ripe",
		Transport: TransportHTTP,
		RemoteFiles: []RemoteFile{
			{Name: "route", Path: "/route.gz"},
		},
		SerialPath: "/CURRENTSERIAL",
		Host:       ts.URL,
	})
	service.httpClient = &http.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := service.getRemoteSerial(ctx)
	assert.Error(t, err)
}
