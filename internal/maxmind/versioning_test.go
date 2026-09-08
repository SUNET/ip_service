package maxmind

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ip_service/pkg/model"

	"github.com/stretchr/testify/assert"
)

// mockKV is a minimal in-memory kvStore for tests.
type mockKV struct {
	remote      map[string]string
	lastChecked map[string]string
}

func (m *mockKV) GetRemoteVersion(_ context.Context, k string) string { return m.remote[k] }
func (m *mockKV) GetLastChecked(_ context.Context, k string) string   { return m.lastChecked[k] }
func (m *mockKV) SetLastChecked(_ context.Context, k string) error {
	if m.lastChecked == nil {
		m.lastChecked = map[string]string{}
	}
	m.lastChecked[k] = "now"
	return nil
}
func (m *mockKV) SetPreviousVersion(_ context.Context, k string) error { return nil }
func (m *mockKV) SetRemoteVersion(_ context.Context, k, v string) error {
	if m.remote == nil {
		m.remote = map[string]string{}
	}
	m.remote[k] = v
	return nil
}

func serviceWithKV(t *testing.T, ts *httptest.Server, kv kvStore) *Service {
	t.Helper()
	s := newServiceWithTestDBs(t)
	s.cfg.IPService.MaxMind = model.MaxMind{
		RemoteURL:     ts.URL,
		ArchiveFormat: "tar.gz",
		BaseFolder:    t.TempDir(),
	}
	s.kvStore = kv
	return s
}

func TestDBMeta_DownloadingInProgress(t *testing.T) {
	m := DBMeta{model.MaxmindDBTypeASN: &DBObject{}}
	assert.False(t, m.IsDownloadingInProgresses(model.MaxmindDBTypeASN))

	m.DownloadInProgress(model.MaxmindDBTypeASN)
	assert.True(t, m.IsDownloadingInProgresses(model.MaxmindDBTypeASN))

	m.DownloadingDone(model.MaxmindDBTypeASN)
	assert.False(t, m.IsDownloadingInProgresses(model.MaxmindDBTypeASN))
}

func TestGetRemoteVersion_HTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("last-modified", "Thu, 01 Sep 2022 18:54:52 GMT")
		w.WriteHeader(200)
	}))
	defer ts.Close()

	s := serviceWithKV(t, ts, &mockKV{})
	v, err := s.getRemoteVersion(t.Context(), model.MaxmindDBTypeASN)
	assert.NoError(t, err)
	assert.NotEmpty(t, v)
}

func TestGetRemoteVersion_429(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	s := serviceWithKV(t, ts, &mockKV{})
	_, err := s.getRemoteVersion(t.Context(), model.MaxmindDBTypeASN)
	assert.Error(t, err)
}

func TestGetRemoteVersion_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	s := serviceWithKV(t, ts, &mockKV{})
	_, err := s.getRemoteVersion(t.Context(), model.MaxmindDBTypeASN)
	assert.Error(t, err)
}

func TestCompareVersion_NewVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("last-modified", "Thu, 01 Sep 2022 18:54:52 GMT")
		w.WriteHeader(200)
	}))
	defer ts.Close()

	s := serviceWithKV(t, ts, &mockKV{})
	got, err := s.compareVersion(t.Context(), model.MaxmindDBTypeASN)
	assert.NoError(t, err)
	assert.True(t, got)
}

func TestCompareVersion_SameVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("last-modified", "Thu, 01 Sep 2022 18:54:52 GMT")
		w.WriteHeader(200)
	}))
	defer ts.Close()

	kv := &mockKV{remote: map[string]string{model.MaxmindDBTypeASN: ""}}
	s := serviceWithKV(t, ts, kv)
	// Prime the store with the current "remote" value so compareVersion returns false.
	_ = kv.SetRemoteVersion(t.Context(), model.MaxmindDBTypeASN, "")
	// First call: sets remote version.
	_, _ = s.compareVersion(t.Context(), model.MaxmindDBTypeASN)
	// Second call: remote and local now match.
	got, err := s.compareVersion(t.Context(), model.MaxmindDBTypeASN)
	assert.NoError(t, err)
	assert.False(t, got)
}

func TestCheckNewDBVersion_ErrorReturnsFalse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	s := serviceWithKV(t, ts, &mockKV{})
	assert.False(t, s.checkNewDBVersion(t.Context(), model.MaxmindDBTypeASN))
}

func TestClose(t *testing.T) {
	s := &Service{quitChan: make(chan struct{}, 1), Log: mustSimpleLog()}
	assert.NoError(t, s.Close(t.Context()))
}
