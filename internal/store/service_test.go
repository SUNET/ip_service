package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ip_service/pkg/logger"
	"ip_service/pkg/model"

	"github.com/stretchr/testify/assert"
)

func mockNew(t *testing.T, dir string) *Service {
	dbPath := filepath.Join(dir, "store")
	cfg := &model.Cfg{
		Store: struct {
			File struct {
				Path string "yaml:\"path\""
			} "yaml:\"file\""
		}{
			File: struct {
				Path string "yaml:\"path\""
			}{
				Path: dbPath,
			},
		},
	}

	s, err := New(context.TODO(), cfg, logger.New("test", false).New("test"))
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

func TestGetWithRemovedSourceFile(t *testing.T) {
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

			assert.NoError(t, os.Remove(filepath.Join(tempDir, "store", "key")))

			assert.Equal(t, "", service.KV.Get(context.TODO(), "key"))

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
