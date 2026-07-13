package rpsl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindNetwork(t *testing.T) {
	tts := []struct {
		name          string
		networkObject *Object
		ip            string
		want          bool
	}{
		{
			name: "SUNET network - inside",
			networkObject: &Object{
				Network: "37.156.192.0/24",
			},
			ip:   "37.156.192.0",
			want: true,
		},
		{
			name: "SUNET network - outside",
			networkObject: &Object{
				Network: "37.156.192.0/7",
			},
			ip:   "37.156.192.2",
			want: true,
		},
		{
			name: "SUNET network - outside",
			networkObject: &Object{
				Network: "37.156.192.0/7",
			},
			ip:   "38.255.255.254",
			want: false,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.networkObject.FindNetwork(context.TODO(), tt.ip)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
