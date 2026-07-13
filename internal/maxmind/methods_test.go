package maxmind

import (
	"context"
	"fmt"
	"ip_service/pkg/model"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockConfig(t *testing.T, storePath, remoteURL string) *model.Cfg {
	cfg := &model.Cfg{
		IPService: &model.IPService{
			APIServer: model.APIServer{
				Addr: "",
			},
			Production: false,
			Log:        model.Log{},
			MaxMind: model.MaxMind{
				AutomaticUpdate:   true,
				UpdatePeriodicity: 100,
				Username:          "testUser",
				Password:          "testPassword",
				Enterprise:        false,
				RetryCounter:      0,
				DB: map[string]model.MaxMindDB{
					model.MaxmindDBTypeASN: {
						FilePath: filepath.Join(storePath, "GeoLite2-asn.mmdb"),
					},
					model.MaxmindDBTypeCity: {
						FilePath: filepath.Join(storePath, "GeoLite2-city.mmdb"),
					},
				},
				BaseFolder:    storePath,
				RemoteURL:     remoteURL,
				ArchiveFormat: "tar.gz",
			},
			Store: model.Store{
				File: model.FileStorage{
					Path: storePath,
				},
			},
			Tracing: model.Tracing{},
		},
	}

	fmt.Println("Mock config", cfg.IPService.MaxMind.RemoteURL, cfg.IPService.MaxMind.BaseFolder)

	return cfg
}

func TestInitial(t *testing.T) {
	tts := []struct {
		name    string
		dbTypes []string
	}{
		{
			name: "empty folder-city",
			dbTypes: []string{
				model.MaxmindDBTypeASN,
				model.MaxmindDBTypeCity,
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()

			mux.HandleFunc("GET /download/", func(w http.ResponseWriter, r *http.Request) {
				t.Log("XXXXXXX Inside of mockServer")

				q := r.URL.Query()
				lic := q.Get("license_key")
				suffix := q.Get("suffix")
				eid := q.Get("edition_id")

				var (
					file []byte = nil
					err  error
				)

				switch eid {
				case "GeoLite2-City":
					file, err = os.ReadFile("./testdata/GeoLite2-City.tar.gz")
					assert.NoError(t, err)
				case "GeoLite2-ASN":
					file, err = os.ReadFile("./testdata/GeoLite2-ASN.tar.gz")
					assert.NoError(t, err)
				default:
					t.Log("no edition_id match found")
					t.FailNow()
				}

				t.Log("XXXXXXX Inside of mockServer", "license_key", lic, "suffix", suffix, "edition_id", eid)
				w.Header().Set("last-modified", latestTS)
				w.Header().Set("Content-Type", "application/gzip")
				w.WriteHeader(200)

				_, err = w.Write(file)
				assert.NoError(t, err)
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			tempDir := t.TempDir()

			_ = mockService(t, tempDir, server.URL, false)

			time.Sleep(1 * time.Second)

			for _, dbType := range tt.dbTypes {
				for _, fileType := range []string{"mmdb"} {
					fileName := fmt.Sprintf("GeoLite2-%s.%s", dbType, fileType)
					s, err := os.Stat(filepath.Join(tempDir, fileName))
					assert.NoError(t, err)

					if s != nil {
						t.Log("fileName", s.Name(), "size", s.Size())
					}
				}
			}
		})
	}
}

func TestOpenDB(t *testing.T) {
	type have struct {
		dbType string
	}
	tts := []struct {
		name string
		have have
		want string
	}{
		{
			name: "1-ASN-OK",
			have: have{
				dbType: model.MaxmindDBTypeASN,
			},
			want: "",
		},
		{
			name: "2-City-OK",
			have: have{
				dbType: model.MaxmindDBTypeCity,
			},
			want: "",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {

			mux := http.NewServeMux()

			mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("XXXXXXX Inside of mockServer")
				w.Header().Set("last-modified", latestTS)
				w.Header().Set("Content-Type", "application/gzip")
				w.WriteHeader(200)
			})

			server := httptest.NewServer(mux)

			tempDir := t.TempDir()

			service := mockService(t, tempDir, server.URL, false)

			err := service.loadDB(context.TODO(), tt.have.dbType)
			assert.NoError(t, err)
		})
	}
}
