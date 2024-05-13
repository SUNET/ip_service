package configuration

import (
	"context"
	"fmt"
	"ip_service/pkg/logger"
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
					HTTPProxy:  "",
					Log:        model.Log{},
					MaxMind: model.MaxMind{
						AutomaticUpdate:   true,
						UpdatePeriodicity: 10,
						LicenseKey:        "",
						Enterprise:        false,
						RetryCounter:      0,
						DB: map[string]model.MaxMindDB{
							"asn": {
								FilePath: "./db/GeoLite2-asn.mmdb",
							},
							"city": {
								FilePath: "./db/GeoLite2-city.mmdb",
							},
						},
					},
					Store: model.Store{
						File: model.FileStorage{
							Path: "kv_store",
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
			path := fmt.Sprintf("../../config.yaml")

			if tt.setEnvVariable {
				err := os.Setenv("CONFIG_YAML", path)
				assert.NoError(t, err)
			}

			testLog := logger.NewSimple("test")
			got, err := Parse(context.TODO(), testLog)
			assert.NoError(t, err)

			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(model.MaxMind{}, "LicenseKey")); diff != "" {
				t.Errorf("diff: %s", diff)
			}

		})
	}

}
