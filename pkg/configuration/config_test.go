package configuration

import (
	"context"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/model"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {

	tts := []struct {
		name           string
		setEnvVariable bool
		want           *model.Cfg
	}{
		{
			name:           "OK",
			setEnvVariable: true,
			want: &model.Cfg{
				IPService: &model.IPService{
					APIServer: model.APIServer{
						Addr: ":8080",
					},
					Production: false,
					Log:        model.Log{},
					Radb: model.Radb{
						FilePath: "/var/lib/ip_service/radb",
					},
					RIPE: model.RIPE{
						FilePath: "/var/lib/ip_service/ripe",
					},
					MaxMind: model.MaxMind{
						AutomaticUpdate: true,
						Enterprise:      false,
						Username:        "668041",
						ArchiveFormat:   "tar.gz",
						BaseFolder:      "/var/lib/ip_service/maxmind",
						RemoteURL:       "https://download.maxmind.com/geoip/databases",
						RetryCounter:    0,
						DB: map[string]model.MaxMindDB{
							model.MaxmindDBTypeASN: {
								FilePath: "/var/lib/ip_service/maxmind/GeoLite2-ASN.mmdb",
							},
							model.MaxmindDBTypeCity: {
								FilePath: "/var/lib/ip_service/maxmind/GeoLite2-City.mmdb",
							},
						},
					},
					Store: model.Store{
						File: model.FileStorage{
							Path: "/var/lib/ip_service/kv_store",
						},
					},
					Tracing: model.Tracing{
						Addr: "ip_service_jaeger:4318",
					},
				},
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			path := "../../configfiles/ip_service.yaml"

			if tt.setEnvVariable {
				err := os.Setenv("CONFIG_YAML", path)
				assert.NoError(t, err)
			}

			testLog := logger.NewSimple("test")
			got, err := Parse(context.TODO(), testLog)
			assert.NoError(t, err)

			ignoreLicKey := cmpopts.IgnoreFields(model.MaxMind{}, "Password")
			ignoreUpdatePeriodicity := cmpopts.IgnoreFields(model.MaxMind{}, "UpdatePeriodicity")

			if diff := cmp.Diff(tt.want, got, ignoreLicKey, ignoreUpdatePeriodicity); diff != "" {
				t.Errorf("diff: %s", diff)
			}

		})
	}

}
