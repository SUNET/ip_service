package rpsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterClassRemoveBlankRecords(t *testing.T) {
	rc := RouterClass{
		"192.168.0.0/16": ASN{"AS1": &Object{Network: "192.168.0.0/16"}},
		"":               ASN{"AS2": &Object{}},
		"10.0.0.0/8":     ASN{"AS3": &Object{Network: "10.0.0.0/8"}},
	}

	rc.removeBlankRecords(t.Context())

	assert.Len(t, rc, 2)
	_, ok := rc[""]
	assert.False(t, ok)
	_, ok = rc["192.168.0.0/16"]
	assert.True(t, ok)
	_, ok = rc["10.0.0.0/8"]
	assert.True(t, ok)
}

func TestRouterClassOpinionatedMerge_IPv6(t *testing.T) {
	tts := []struct {
		name string
		r1   RouterClass
		r2   RouterClass
		want RouterClass
	}{
		{
			name: "IPv6 - r2 takes precedence",
			r1: RouterClass{
				"2001:db8::/32": ASN{
					"AS1000": &Object{Network: "2001:db8::/32", Origin: "AS1000"},
				},
			},
			r2: RouterClass{
				"2001:db8::/32": ASN{
					"AS1000": &Object{Network: "2001:db8::/32", Origin: "AS1000"},
				},
			},
			want: RouterClass{
				"2001:db8::/32": ASN{
					"AS1000": &Object{Network: "2001:db8::/32", Origin: "AS1000"},
				},
			},
		},
		{
			name: "IPv6 - new ASN added to existing network",
			r1: RouterClass{
				"2001:67c:2564::/48": ASN{
					"AS1653": &Object{Network: "2001:67c:2564::/48", Origin: "AS1653"},
					"AS9999": &Object{Network: "2001:67c:2564::/48", Origin: "AS9999"},
				},
			},
			r2: RouterClass{
				"2001:67c:2564::/48": ASN{
					"AS1653": &Object{Network: "2001:67c:2564::/48", Origin: "AS1653"},
				},
			},
			want: RouterClass{
				"2001:67c:2564::/48": ASN{
					"AS1653": &Object{Network: "2001:67c:2564::/48", Origin: "AS1653"},
					"AS9999": &Object{Network: "2001:67c:2564::/48", Origin: "AS9999"},
				},
			},
		},
		{
			name: "IPv6 - new network from r1",
			r1: RouterClass{
				"2001:db9::/32": ASN{
					"AS2000": &Object{Network: "2001:db9::/32", Origin: "AS2000"},
				},
			},
			r2: RouterClass{},
			want: RouterClass{
				"2001:db9::/32": ASN{
					"AS2000": &Object{Network: "2001:db9::/32", Origin: "AS2000"},
				},
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RouterClassOpinionatedMerge(t.Context(), tt.r1, tt.r2)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
