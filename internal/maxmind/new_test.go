package maxmind

import (
	"os"
	"path/filepath"
	"testing"

	"ip_service/internal/store"
	"ip_service/pkg/model"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/stretchr/testify/assert"
)

// TestNew_WithArchivePresent exercises the New constructor by pre-placing an
// archive so initial() succeeds without needing to reach a remote.
func TestNew_WithArchivePresent(t *testing.T) {
	tmp := t.TempDir()

	// Copy both archives so both ASN and City initialize.
	for _, dbType := range []string{model.MaxmindDBTypeASN, model.MaxmindDBTypeCity} {
		src := filepath.Join("testdata", "GeoLite2-"+dbType+".tar.gz")
		dst := filepath.Join(tmp, "GeoLite2-"+dbType+".tar.gz")
		copyFile(t, src, dst)
	}

	ctx := t.Context()
	tp, err := trace.NewForTesting(ctx, "t", logger.NewSimple("t"))
	assert.NoError(t, err)

	cfg := &model.Cfg{
		IPService: &model.IPService{
			Store: model.Store{File: model.FileStorage{Path: tmp}},
			MaxMind: model.MaxMind{
				BaseFolder:        tmp,
				RemoteURL:         "http://127.0.0.1:1/download/",
				ArchiveFormat:     "tar.gz",
				UpdatePeriodicity: 3600,
			},
		},
	}
	st, err := store.New(ctx, cfg, tp, logger.NewSimple("store"))
	assert.NoError(t, err)

	s, err := New(ctx, cfg, st, tp, logger.NewSimple("maxmind"))
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.NotNil(t, s.DBASN)
	assert.NotNil(t, s.DBCity)

	assert.NoError(t, s.Close(ctx))

	// Verify the extracted files ended up in the expected location.
	_, err = os.Stat(filepath.Join(tmp, "GeoLite2-ASN.mmdb"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmp, "GeoLite2-City.mmdb"))
	assert.NoError(t, err)
}
