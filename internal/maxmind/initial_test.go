package maxmind

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"ip_service/internal/store"
	"ip_service/pkg/model"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	assert.NoError(t, err)
	defer in.Close()
	out, err := os.Create(dst)
	assert.NoError(t, err)
	defer out.Close()
	_, err = io.Copy(out, in)
	assert.NoError(t, err)
}

func newInitService(t *testing.T, storePath string) *Service {
	t.Helper()
	ctx := t.Context()
	tp, err := trace.NewForTesting(ctx, "t", logger.NewSimple("t"))
	assert.NoError(t, err)

	cfg := &model.Cfg{
		IPService: &model.IPService{
			Store: model.Store{File: model.FileStorage{Path: storePath}},
			MaxMind: model.MaxMind{
				BaseFolder:    storePath,
				RemoteURL:     "http://127.0.0.1:0/download/",
				ArchiveFormat: "tar.gz",
			},
		},
	}
	st, err := store.New(ctx, cfg, tp, logger.NewSimple("store"))
	assert.NoError(t, err)

	return &Service{
		probeStore:   &model.StatusProbeStore{},
		cfg:          cfg,
		Log:          logger.NewSimple("maxmind"),
		TP:           tp,
		kvStore:      st.KV,
		reloadChan:   make(chan string, 10),
		downloadChan: make(chan string, 10),
		updateChan:   make(chan string, 10),
		initialChan:  make(chan string, 10),
		DBMeta: DBMeta{
			model.MaxmindDBTypeASN:  {rateLimit: *rate.NewLimiter(rate.Every(1), 1)},
			model.MaxmindDBTypeCity: {rateLimit: *rate.NewLimiter(rate.Every(1), 1)},
		},
	}
}

// TestInitial_ArchivePresent copies an archive into BaseFolder and verifies
// initial extracts and loads it (covering unTarV3 + loadDB).
func TestInitial_ArchivePresent(t *testing.T) {
	tmp := t.TempDir()
	copyFile(t, "./testdata/GeoLite2-ASN.tar.gz", filepath.Join(tmp, "GeoLite2-ASN.tar.gz"))

	s := newInitService(t, tmp)

	err := s.initial(t.Context(), model.MaxmindDBTypeASN)
	assert.NoError(t, err)

	// After initial, the mmdb should exist and DBASN should be loaded.
	_, err = os.Stat(filepath.Join(tmp, "GeoLite2-ASN.mmdb"))
	assert.NoError(t, err)
	assert.NotNil(t, s.DBASN)
}

func TestUnTarV3_MissingArchive(t *testing.T) {
	s := newInitService(t, t.TempDir())
	err := s.unTarV3(t.Context(), model.MaxmindDBTypeASN)
	assert.Error(t, err)
}
