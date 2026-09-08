package whois

import (
	"context"
	"testing"
	"time"

	"ip_service/internal/lctree"
	"ip_service/internal/store"
	"ip_service/pkg/model"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/stretchr/testify/assert"
)

// TestNew_NetworkFailure exercises New. It relies on RIPE/RADB update calls
// failing when the network is unreachable — Update errors are only logged,
// so New still returns a valid Service. Skipped in short mode because the
// downloader may attempt (and slowly time out) DNS lookups.
func TestNew_NetworkFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("network-dependent: skipping in -short mode")
	}

	tmp := t.TempDir()
	ctx := t.Context()
	tp, err := trace.NewForTesting(ctx, "t", logger.NewSimple("t"))
	assert.NoError(t, err)

	cfg := &model.Cfg{
		IPService: &model.IPService{
			Store: model.Store{File: model.FileStorage{Path: tmp}},
			Radb:  model.Radb{FilePath: tmp + "/radb.db"},
			RIPE:  model.RIPE{FilePath: tmp + "/ripe.db"},
		},
	}
	st, err := store.New(ctx, cfg, tp, logger.NewSimple("store"))
	assert.NoError(t, err)

	tree := lctree.New(logger.NewSimple("lctree"))

	// Fail fast if this ends up racing external network.
	deadlineCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	svc, err := New(deadlineCtx, cfg, tree, st, logger.NewSimple("whois"))
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	assert.NoError(t, svc.Close(ctx))
}
