package store

import (
	"context"
	"path/filepath"
	"testing"

	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"ip_service/pkg/trace"

	"github.com/stretchr/testify/assert"
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
			MaxMind:    model.MaxMind{},
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

func mockNew(t *testing.T, dir string) *Service {
	ctx := context.TODO()
	storePath := filepath.Join(dir, "store")
	cfg := mockConfig(t, storePath)

	tp, err := trace.New(ctx, cfg, logger.NewSimple("test"), "test", "test")
	assert.NoError(t, err)

	s, err := New(ctx, cfg, tp, logger.NewSimple("test"))
	assert.NoError(t, err)
	assert.NoError(t, s.KV.Set(context.TODO(), "key", "value"))

	return s
}

func TestGet(t *testing.T) {
	tts := []struct {
		name string
	}{
		{
			name: "OK",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			service := mockNew(t, tempDir)

			assert.Equal(t, "value", service.KV.Get(context.TODO(), "key"))

		})
	}
}

func TestWrite(t *testing.T) {
	tts := []struct {
		name string
		have string
		want string
	}{
		{
			name: "OK",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			// Test function
			tempDir := t.TempDir()
			s := mockNew(t, tempDir)
			err := s.KV.Set(context.TODO(), "mura/mura", "muuu")
			assert.NoError(t, err)

			got := s.KV.Get(context.TODO(), "mura/mura")
			assert.Equal(t, "muuu", got)

			err = s.KV.Set(context.TODO(), "mura/mura/mura", "kalle")
			assert.NoError(t, err)

			got = s.KV.Get(context.TODO(), "mura/mura/mura")
			assert.Equal(t, "kalle", got)

		})
	}
}

func TestRemoteVersion(t *testing.T) {
	tts := []struct {
		name string
		have string
		want string
	}{
		{
			name: "OK",
			want: "now",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			s := mockNew(t, tempDir)
			err := s.KV.SetRemoteVersion(context.TODO(), "asn", "now")
			assert.NoError(t, err)

			got := s.KV.GetRemoteVersion(context.TODO(), "asn")
			assert.Equal(t, tt.want, got)
		})
	}
}
