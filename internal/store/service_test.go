package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/model"
	"github.com/SUNET/vc/pkg/trace"

	"github.com/stretchr/testify/assert"
)

func mockConfig(storePath string) *model.Cfg {
	cfg := &model.Cfg{
		IPService: &model.IPService{
			APIServer: model.APIServer{
				Addr: "",
			},
			Production: false,
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
	cfg := mockConfig(storePath)

	tp, err := trace.NewForTesting(ctx, "test", logger.NewSimple("test"))
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
	type have struct {
		key   string
		value string
	}
	tts := []struct {
		name string
		have have
		want string
	}{
		{
			name: "1",
			have: have{
				key:   "key/key",
				value: "value1",
			},
			want: "value1",
		},
		{
			name: "2",
			have: have{
				key:   "key/key/key",
				value: "value2",
			},
			want: "value2",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			s := mockNew(t, tempDir)
			err := s.KV.Set(context.TODO(), tt.have.key, tt.have.value)
			assert.NoError(t, err)

			got := s.KV.Get(context.TODO(), tt.have.key)
			assert.Equal(t, tt.want, got)
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
