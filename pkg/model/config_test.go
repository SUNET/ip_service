package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	tts := []struct {
		name string
		have *Cfg
		want string
	}{
		{
			name: "1-OK",
			have: &Cfg{
				IPService: &IPService{
					MaxMind: MaxMind{
						AutomaticUpdate:   false,
						UpdatePeriodicity: 0,
						BaseFolder:        "",
						Enterprise:        false,
						RetryCounter:      0,
						DB:                map[string]MaxMindDB{},
						RemoteURL:         "https://example.com/geoip/databases",
						ArchiveFormat:     "tar.gz",
					},
				},
			},
			want: "https://example.com/geoip/databases/GeoLite2-ASN/download?suffix=tar.gz",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.IPService.MaxMind.URL(MaxmindDBTypeASN)
			assert.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaxMindDBPresent(t *testing.T) {
	tts := []struct {
		have   *Cfg
		want   bool
		dbType string
	}{
		{
			dbType: "asn",
			have:   &Cfg{IPService: &IPService{MaxMind: MaxMind{BaseFolder: t.TempDir()}}},
			want:   true,
		},
		{
			dbType: "city",
			have:   &Cfg{IPService: &IPService{MaxMind: MaxMind{BaseFolder: t.TempDir()}}},
			want:   true,
		},
	}

	for _, tt := range tts {
		t.Run(tt.dbType, func(t *testing.T) {
			dbFilePath := tt.have.IPService.MaxMind.DBFilePath(tt.dbType)
			_, err := os.Create(filepath.Clean(dbFilePath))
			assert.NoError(t, err)
			defer os.Remove(dbFilePath)

			assert.Equal(t, tt.want, tt.have.IPService.MaxMind.IsDBPresent(tt.dbType))

			archiveFilePath := tt.have.IPService.MaxMind.ArchiveFilePath(tt.dbType)
			_, err = os.Create(filepath.Clean(archiveFilePath))
			assert.NoError(t, err)
			defer os.Remove(dbFilePath)

			assert.Equal(t, tt.want, tt.have.IPService.MaxMind.IsArchivePresent(tt.dbType))

		})
	}
}

func TestMaxMindDBFilepath(t *testing.T) {
	tts := []struct {
		have   *Cfg
		dbType string
		want   string
	}{
		{
			dbType: "asn",
			have:   &Cfg{IPService: &IPService{MaxMind: MaxMind{BaseFolder: "/tmp/db"}}},
			want:   "/tmp/db/GeoLite2-asn.mmdb",
		},
		{
			dbType: "city",
			have:   &Cfg{IPService: &IPService{MaxMind: MaxMind{BaseFolder: "/tmp/db"}}},
			want:   "/tmp/db/GeoLite2-city.mmdb",
		},
	}

	for _, tt := range tts {
		t.Run(tt.dbType, func(t *testing.T) {
			got := tt.have.IPService.MaxMind.DBFilePath(tt.dbType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMaxMindArchiveFilepath(t *testing.T) {
	tts := []struct {
		have   *Cfg
		dbType string
		want   string
	}{
		{
			dbType: "asn",
			have:   &Cfg{IPService: &IPService{MaxMind: MaxMind{BaseFolder: "/tmp/db"}}},
			want:   "/tmp/db/GeoLite2-asn.tar.gz",
		},
		{
			dbType: "city",
			have:   &Cfg{IPService: &IPService{MaxMind: MaxMind{BaseFolder: "/tmp/db"}}},
			want:   "/tmp/db/GeoLite2-city.tar.gz",
		},
	}

	for _, tt := range tts {
		t.Run(tt.dbType, func(t *testing.T) {
			got := tt.have.IPService.MaxMind.ArchiveFilePath(tt.dbType)
			assert.Equal(t, tt.want, got)
		})
	}
}
