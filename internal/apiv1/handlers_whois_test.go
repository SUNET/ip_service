package apiv1

import (
	"ip_service/internal/lctree"
	"ip_service/internal/whois"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/rpsl"
	"github.com/SUNET/vc/pkg/trace"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockWhoisClient(t *testing.T, routerClass rpsl.RouterClass) *Client {
	t.Helper()

	log := logger.NewSimple("testing")
	tracer, err := trace.NewForTesting(t.Context(), "test", logger.NewSimple("test"))
	assert.NoError(t, err)

	tree := lctree.New(log)
	err = tree.Build(t.Context(), routerClass)
	assert.NoError(t, err)

	ws := whois.NewTestService(tree, routerClass)

	return &Client{
		log:   log,
		tp:    tracer,
		whois: ws,
	}
}

func TestWhois(t *testing.T) {
	mockObject := &rpsl.Object{
		Network: "2001:db8::/32",
	}

	tts := []struct {
		name        string
		routerClass rpsl.RouterClass
		request     *WhoisRequest
		wantLen     int
		wantErr     bool
	}{
		{
			name: "valid IPv6 - single match",
			routerClass: rpsl.RouterClass{
				"2001:db8::/32": rpsl.ASN{
					"AS64512": mockObject,
				},
			},
			request: &WhoisRequest{IP: "2001:db8::1"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "valid IPv6 - multiple overlapping prefixes",
			routerClass: rpsl.RouterClass{
				"2001:db8::/32": rpsl.ASN{
					"AS64512": mockObject,
				},
				"2001:db8:1::/48": rpsl.ASN{
					"AS64513": mockObject,
				},
			},
			request: &WhoisRequest{IP: "2001:db8:1::1"},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "valid IPv6 - no match",
			routerClass: rpsl.RouterClass{
				"2001:db8::/32": rpsl.ASN{
					"AS64512": mockObject,
				},
			},
			request: &WhoisRequest{IP: "2001:db9::1"},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "valid IPv4 - single match",
			routerClass: rpsl.RouterClass{
				"192.168.0.0/16": rpsl.ASN{
					"AS64512": mockObject,
				},
			},
			request: &WhoisRequest{IP: "192.168.1.1"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "valid IPv4 - no match",
			routerClass: rpsl.RouterClass{
				"10.0.0.0/8": rpsl.ASN{
					"AS64512": mockObject,
				},
			},
			request: &WhoisRequest{IP: "172.16.0.1"},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:        "invalid IP address",
			routerClass: rpsl.RouterClass{},
			request:     &WhoisRequest{IP: "not-an-ip"},
			wantLen:     0,
			wantErr:     true,
		},
		{
			name:        "empty IP address",
			routerClass: rpsl.RouterClass{},
			request:     &WhoisRequest{IP: ""},
			wantLen:     0,
			wantErr:     true,
		},
		{
			name: "IPv6 most specific prefix returned",
			routerClass: rpsl.RouterClass{
				"2001:67c:2564::/48": rpsl.ASN{
					"AS1653": mockObject,
				},
				"2001:67c::/32": rpsl.ASN{
					"AS1653": mockObject,
				},
			},
			request: &WhoisRequest{IP: "2001:67c:2564::1"},
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			client := mockWhoisClient(t, tt.routerClass)

			got, err := client.Whois(t.Context(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.wantLen == 0 {
				assert.Nil(t, got)
			} else {
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}
