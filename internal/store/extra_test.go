package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKV_LastCheckedAndPreviousVersion(t *testing.T) {
	dir := t.TempDir()
	s := mockNew(t, dir)

	ctx := t.Context()
	assert.NoError(t, s.KV.SetLastChecked(ctx, "asn"))
	assert.NotEmpty(t, s.KV.GetLastChecked(ctx, "asn"))

	assert.NoError(t, s.KV.SetPreviousVersion(ctx, "asn"))
}

func TestKV_Del(t *testing.T) {
	dir := t.TempDir()
	s := mockNew(t, dir)

	ctx := t.Context()
	assert.NoError(t, s.KV.Set(ctx, "todelete", "v"))
	assert.Equal(t, "v", s.KV.Get(ctx, "todelete"))

	assert.NoError(t, s.KV.Del(ctx, "todelete"))
	assert.Equal(t, "", s.KV.Get(ctx, "todelete"))
}

func TestService_Status_OK(t *testing.T) {
	dir := t.TempDir()
	s := mockNew(t, dir)

	probe := s.Status(t.Context())
	assert.NotNil(t, probe)
	assert.Equal(t, "kv", probe.Name)
	assert.True(t, probe.Healthy)

	// Second call within the cache window returns the cached result.
	probe2 := s.Status(t.Context())
	assert.Same(t, probe, probe2)
}

func TestService_Close(t *testing.T) {
	dir := t.TempDir()
	s := mockNew(t, dir)
	assert.NoError(t, s.Close(t.Context()))
}
