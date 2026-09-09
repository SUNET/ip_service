package apiv1

import (
	"testing"

	"ip_service/internal/lctree"
	"ip_service/internal/maxmind"
	"ip_service/internal/store"
	"ip_service/internal/whois"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/model"
	"ip_service/pkg/rpsl"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := t.Context()
	tp, err := trace.NewForTesting(ctx, "t", logger.NewSimple("t"))
	assert.NoError(t, err)

	c, err := New(ctx, &maxmind.Service{}, &whois.Service{}, &store.Service{}, &model.Cfg{}, tp, logger.NewSimple("t"))
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func mockClientWithWhois(t *testing.T, rc rpsl.RouterClass) *Client {
	t.Helper()
	c := mockClient(t)

	tree := lctree.New(logger.NewSimple("lctree"))
	assert.NoError(t, tree.Build(t.Context(), rc))
	c.whois = whois.NewTestService(tree, rc)
	return c
}

func TestLookUpIP_OK(t *testing.T) {
	c := mockClientWithWhois(t, rpsl.RouterClass{})
	ctx := contexthandler.Add(t.Context(), "request", &contexthandler.RequestContext{
		ClientIP: "127.0.0.1",
	})
	got, err := c.LookUpIP(ctx, &LookUpIPRequest{IP: "89.160.20.112"})
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "89.160.20.112", got.IP)
	assert.NotNil(t, got.PTR)
}

func TestLookUpIP_MissingContext(t *testing.T) {
	c := mockClient(t)
	_, err := c.LookUpIP(t.Context(), &LookUpIPRequest{IP: "89.160.20.112"})
	assert.Error(t, err)
}

func TestStatus(t *testing.T) {
	c := mockClient(t)
	// Provide a working store so probe won't panic.
	dir := t.TempDir()
	stCfg := &model.Cfg{IPService: &model.IPService{Store: model.Store{File: model.FileStorage{Path: dir}}}}
	tp, err := trace.NewForTesting(t.Context(), "t", logger.NewSimple("t"))
	assert.NoError(t, err)
	st, err := store.New(t.Context(), stCfg, tp, logger.NewSimple("store"))
	assert.NoError(t, err)
	c.store = st

	got, err := c.Status(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "ip_service", got.Data.ServiceName)
}
