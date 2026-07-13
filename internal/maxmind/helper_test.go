package maxmind

import (
	"context"
	"fmt"
	"ip_service/internal/store"
	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	latestTS = "Thu, 01 Sep 2022 18:54:52 GMT"
	oldTS    = "Thu, 01 Aug 2022 18:54:52 GMT"
)

func mockService(t *testing.T, storeDir, serverURL string, preOpenDB bool) *Service {
	u, err := url.Parse(fmt.Sprintf("%s/%s/", serverURL, "download"))
	assert.NoError(t, err)

	urlValues := u.Query()

	urlValues.Add("license_key", "testKey")
	urlValues.Add("suffix", "tar.gz")
	urlValues.Add("edition_id", "GeoLite2-city")

	u.RawQuery = urlValues.Encode()

	fmt.Println("urlValues", urlValues, "url", u.String())
	//mockGenericEndpointServer(t, mux, "HEAD", u.String(), nil, 200)
	//defer server.Close()

	tracer, err := trace.NewForTesting(context.TODO(), "test", logger.NewSimple("test"))
	assert.NoError(t, err)

	cfg := mockConfig(t, storeDir, fmt.Sprintf("%s/%s/", serverURL, "download"))

	log, err := logger.New("test", "", false)
	assert.NoError(t, err)

	store, err := store.New(context.TODO(), cfg, tracer, log.New("store"))
	assert.NoError(t, err)

	s, err := New(context.TODO(), cfg, store, tracer, log.New("maxmind"))
	assert.NoError(t, err)

	return s
}
