package rpsl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
)

func TestParse(t *testing.T) {
	tts := []struct {
		name     string
		filePath string
	}{
		{
			name:     "SUNET RADB file",
			filePath: "mix.golden",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rpslService, err := New(ctx)
			assert.NoError(t, err)

			interCount = 0

			err = rpslService.Parse(ctx, "./testdata/"+tt.filePath)
			assert.NoError(t, err)

			got, err := json.Marshal(rpslService.RouterClass)
			assert.NoError(t, err)

			var want RouterClass
			wantByte := golden.Get(t, "router_object_formatted.json")
			err = json.Unmarshal(wantByte, &want)
			assert.NoError(t, err)

			wantJSON, err := json.Marshal(want)
			assert.NoError(t, err)

			assert.JSONEq(t, string(wantJSON), string(got))
		})
	}
}

func TestRouterClassOpinionatedMerge(t *testing.T) {
	tts := []struct {
		name string
		r1   RouterClass
		r2   RouterClass
		want RouterClass
	}{
		{
			name: "nothing to merge",
			r1: RouterClass{
				"192.0.1.0/24": {
					"AS12345": {
						Network: "192.0.1.0/24",
						Origin:  "AS12345",
					},
				},
			},
			r2: RouterClass{
				"192.0.1.0/24": {
					"AS12345": {
						Network: "192.0.1.0/24",
						Origin:  "AS12345",
					},
				},
			},
			want: RouterClass{
				"192.0.1.0/24": {
					"AS12345": {
						Network: "192.0.1.0/24",
						Origin:  "AS12345",
					},
				},
			},
		},
		{
			name: "simple merge",
			r1: RouterClass{
				"192.0.2.0/24": {
					"AS12345": {
						Network: "192.0.2.0/24",
						Origin:  "AS12345",
					},
				},
			},
			r2: RouterClass{
				"192.0.2.0/24": {
					"AS67890": {
						Network: "192.0.2.0/24",
						Origin:  "AS67890",
					},
				},
			},
			want: RouterClass{
				"192.0.2.0/24": {
					"AS12345": {
						Network: "192.0.2.0/24",
						Origin:  "AS12345",
					},
					"AS67890": {
						Network: "192.0.2.0/24",
						Origin:  "AS67890",
					},
				},
			},
		},
		{
			name: "add new network",
			r1: RouterClass{
				"192.0.5.0/24": {
					"AS12345": {
						Network: "192.0.5.0/24",
						Origin:  "AS12345",
					},
				},
			},
			r2: RouterClass{
				"192.0.3.0/24": {
					"AS67890": {
						Network: "192.0.3.0/24",
						Origin:  "AS67890",
					},
				},
			},
			want: RouterClass{
				"192.0.5.0/24": {
					"AS12345": {
						Network: "192.0.5.0/24",
						Origin:  "AS12345",
					},
				},
				"192.0.3.0/24": {
					"AS67890": {
						Network: "192.0.3.0/24",
						Origin:  "AS67890",
					},
				},
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			got, err := RouterClassOpinionatedMerge(ctx, tt.r1, tt.r2)
			assert.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
