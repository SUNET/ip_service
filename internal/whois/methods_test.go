package whois

import (
	"net/netip"
	"testing"

	"ip_service/internal/lctree"
	"ip_service/pkg/rpsl"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func newTestServiceWith(t *testing.T, rc rpsl.RouterClass) *Service {
	t.Helper()
	tree := lctree.New(logger.NewSimple("whois-test"))
	assert.NoError(t, tree.Build(t.Context(), rc))
	return NewTestService(tree, rc)
}

func TestQueryIP_Match(t *testing.T) {
	obj := &rpsl.Object{Network: "2001:db8::/32"}
	rc := rpsl.RouterClass{
		"2001:db8::/32": rpsl.ASN{"AS1": obj},
	}
	s := newTestServiceWith(t, rc)

	asn, err := s.QueryIP(t.Context(), "2001:db8::1")
	assert.NoError(t, err)
	assert.NotNil(t, asn)
	assert.Contains(t, asn, "AS1")
}

func TestQueryIP_NoMatch(t *testing.T) {
	s := newTestServiceWith(t, rpsl.RouterClass{
		"2001:db8::/32": rpsl.ASN{"AS1": &rpsl.Object{}},
	})

	asn, err := s.QueryIP(t.Context(), "2001:db9::1")
	assert.NoError(t, err)
	assert.Nil(t, asn)
}

func TestQueryIP_InvalidIP(t *testing.T) {
	s := newTestServiceWith(t, rpsl.RouterClass{})
	_, err := s.QueryIP(t.Context(), "not-an-ip")
	assert.Error(t, err)
}

func TestQueryIPAll_MultiplePrefixes(t *testing.T) {
	rc := rpsl.RouterClass{
		"2001:db8::/32":   rpsl.ASN{"AS1": &rpsl.Object{}},
		"2001:db8:1::/48": rpsl.ASN{"AS2": &rpsl.Object{}},
	}
	s := newTestServiceWith(t, rc)

	list, err := s.QueryIPAll(t.Context(), netip.MustParseAddr("2001:db8:1::1"))
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestQueryIPAll_NoMatch(t *testing.T) {
	s := newTestServiceWith(t, rpsl.RouterClass{
		"192.168.0.0/16": rpsl.ASN{"AS1": &rpsl.Object{}},
	})
	list, err := s.QueryIPAll(t.Context(), netip.MustParseAddr("10.0.0.1"))
	assert.NoError(t, err)
	assert.Nil(t, list)
}
