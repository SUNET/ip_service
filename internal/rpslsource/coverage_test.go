package rpslsource

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// A minimal RPSL "route" record that pkg/rpsl.Parse can consume.
const minimalRouteObject = "route:  10.0.0.0/8\norigin: AS64500\n\n"

func gzBytes(t *testing.T, s string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(s))
	assert.NoError(t, err)
	assert.NoError(t, gw.Close())
	return buf.Bytes()
}

func newHTTPService(t *testing.T, ts *httptest.Server, name string, addEOF bool) *Service {
	t.Helper()
	s := mockService(t, Config{
		Name:      name,
		Transport: TransportHTTP,
		RemoteFiles: []RemoteFile{
			{Name: "route", Path: "/route.gz"},
		},
		SerialPath:   "/CURRENTSERIAL",
		Host:         ts.URL,
		AddEOFMarker: addEOF,
	})
	// Simple client to avoid retryablehttp overhead in tests.
	s.httpClient = &http.Client{Timeout: 5 * time.Second}
	// Loose rate limit for tests.
	s.rateLimit = *rate.NewLimiter(rate.Every(time.Millisecond), 100)
	return s
}

func TestArchiveAndLocalPath(t *testing.T) {
	s := mockService(t, Config{Name: "acme", Transport: TransportHTTP})
	assert.Equal(t, "/tmp/acme_db_route.gz", s.archivePath("route"))
	assert.True(t, strings.HasSuffix(s.localFilePath("route"), "acme.db_route.txt"))
}

func TestGetRemoteSerial_UnknownTransport(t *testing.T) {
	s := mockService(t, Config{Name: "x", Transport: TransportType(99)})
	_, err := s.getRemoteSerial(t.Context())
	assert.Error(t, err)
}

func TestDownloadArchive_UnknownTransport(t *testing.T) {
	s := mockService(t, Config{Name: "x", Transport: TransportType(99)})
	err := s.downloadArchive(t.Context(), RemoteFile{Name: "route", Path: "/foo"})
	assert.Error(t, err)
}

func TestDownloadArchive_RateLimit(t *testing.T) {
	s := mockService(t, Config{Name: "x", Transport: TransportHTTP})
	// Zero-tokens limiter to force rate limit.
	s.rateLimit = *rate.NewLimiter(rate.Every(time.Hour), 0)
	err := s.downloadArchive(t.Context(), RemoteFile{Name: "route", Path: "/foo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
}

func TestDownloadArchiveHTTP_429(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	s := newHTTPService(t, ts, "acme", false)
	err := s.downloadArchiveHTTP(t.Context(), RemoteFile{Name: "route", Path: "/route.gz"})
	assert.Error(t, err)
}

func TestDownloadHTTP_UnzipAndParse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/CURRENTSERIAL":
			_, _ = w.Write([]byte("42"))
		case "/route.gz":
			_, _ = w.Write(gzBytes(t, minimalRouteObject))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	s := newHTTPService(t, ts, "acme", true)

	// Update walks the full HTTP flow: serial -> download -> unzip -> parse.
	updated, err := s.Update(t.Context())
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.Contains(t, s.RPSLRouterClass, "10.0.0.0/8")

	// Serial is now cached; a second update should still succeed via loadFromLocal.
	updated, err = s.Update(t.Context())
	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestUnzip_CorruptGzip(t *testing.T) {
	s := mockService(t, Config{Name: "corrupt", Transport: TransportHTTP})
	arch := s.archivePath("route")
	assert.NoError(t, os.WriteFile(arch, []byte("not gzip"), 0600))
	defer os.Remove(arch)

	err := s.unzip(t.Context(), "route")
	assert.Error(t, err)
}

func TestAddEOFMarker(t *testing.T) {
	s := mockService(t, Config{Name: "eof", Transport: TransportHTTP})
	local := s.localFilePath("route")
	assert.NoError(t, os.WriteFile(local, []byte("route: 10.0.0.0/8\n"), 0600))
	defer os.Remove(local)

	assert.NoError(t, s.addEOFMarker("route"))
	body, err := os.ReadFile(local)
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(string(body), "\n# EOF\n"))
}

func TestCleanupArchive(t *testing.T) {
	s := mockService(t, Config{Name: "clean", Transport: TransportHTTP})
	arch := s.archivePath("route")
	assert.NoError(t, os.WriteFile(arch, []byte("x"), 0600))
	assert.NoError(t, s.cleanupArchive("route"))
	_, err := os.Stat(arch)
	assert.Error(t, err)
}

func TestLoadFromLocal_MissingFile(t *testing.T) {
	s := mockService(t, Config{
		Name:        "missing",
		Transport:   TransportHTTP,
		RemoteFiles: []RemoteFile{{Name: "route", Path: "/route.gz"}},
	})
	assert.False(t, s.loadFromLocal(t.Context()))
}

func TestLoadFromLocal_BadContent(t *testing.T) {
	s := mockService(t, Config{
		Name:        "badcontent",
		Transport:   TransportHTTP,
		RemoteFiles: []RemoteFile{{Name: "route", Path: "/route.gz"}},
	})
	// Point to something outside t.TempDir so it survives after mockService returns.
	local := s.localFilePath("route")
	// Write an unreadable/malformed input (empty file is parseable but yields no records).
	assert.NoError(t, os.WriteFile(local, []byte("route: notavalidprefix\n\n"), 0600))
	defer os.Remove(local)

	// loadFromLocal reports true when Parse succeeds even without valid entries.
	got := s.loadFromLocal(t.Context())
	assert.True(t, got)
}

func TestClose(t *testing.T) {
	s := mockService(t, Config{Name: "close", Transport: TransportHTTP})
	assert.NoError(t, s.Close(t.Context()))
}

func TestUpdate_HTTP_SerialFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := newHTTPService(t, ts, "acme", false)
	updated, err := s.Update(t.Context())
	assert.Error(t, err)
	assert.False(t, updated)
}

func TestUpdate_HTTP_DownloadFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/CURRENTSERIAL":
			_, _ = w.Write([]byte("newserial"))
		case "/route.gz":
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	s := newHTTPService(t, ts, "acme", false)
	updated, err := s.Update(t.Context())
	assert.Error(t, err)
	assert.False(t, updated)
}
