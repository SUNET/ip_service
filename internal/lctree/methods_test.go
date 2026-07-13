package lctree

import (
	"context"
	"fmt"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/rpsl"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	log := logger.NewSimple("lctree_test")
	return New(log)
}

func TestBuild(t *testing.T) {
	s := newTestService(t)

	rc := rpsl.RouterClass{
		"2001:db8::/32":      nil,
		"2001:db8:1::/48":    nil,
		"192.168.0.0/16":     nil,
		"192.168.1.0/24":     nil,
		"10.0.0.0/8":         nil,
		"ffff:db8:0:1::/64":  nil,
	}

	err := s.Build(context.Background(), rc)
	assert.NoError(t, err)

	v4, v6 := s.CountTags()
	assert.Equal(t, 3, v4)
	assert.Equal(t, 3, v6)
}

func TestFindDeepestTag_IPv6(t *testing.T) {
	s := newTestService(t)

	rc := rpsl.RouterClass{
		"2001:db8::/32":   nil,
		"2001:db8:1::/48": nil,
	}

	err := s.Build(context.Background(), rc)
	assert.NoError(t, err)

	// IP in the /48 should return the more specific prefix
	tag, found := s.FindDeepestTag(netip.MustParseAddr("2001:db8:1::1"))
	assert.True(t, found)
	assert.Equal(t, "2001:db8:1::/48", tag)

	// IP in the /32 but not /48 should return the broader prefix
	tag, found = s.FindDeepestTag(netip.MustParseAddr("2001:db8:2::1"))
	assert.True(t, found)
	assert.Equal(t, "2001:db8::/32", tag)

	// IP not in any prefix
	_, found = s.FindDeepestTag(netip.MustParseAddr("2002:db8::1"))
	assert.False(t, found)
}

func TestFindDeepestTag_IPv4(t *testing.T) {
	s := newTestService(t)

	rc := rpsl.RouterClass{
		"192.168.0.0/16": nil,
		"192.168.1.0/24": nil,
		"10.0.0.0/8":     nil,
	}

	err := s.Build(context.Background(), rc)
	assert.NoError(t, err)

	tag, found := s.FindDeepestTag(netip.MustParseAddr("192.168.1.100"))
	assert.True(t, found)
	assert.Equal(t, "192.168.1.0/24", tag)

	tag, found = s.FindDeepestTag(netip.MustParseAddr("192.168.2.1"))
	assert.True(t, found)
	assert.Equal(t, "192.168.0.0/16", tag)

	tag, found = s.FindDeepestTag(netip.MustParseAddr("10.1.2.3"))
	assert.True(t, found)
	assert.Equal(t, "10.0.0.0/8", tag)

	_, found = s.FindDeepestTag(netip.MustParseAddr("172.16.0.1"))
	assert.False(t, found)
}

func TestFindTags_AllMatching(t *testing.T) {
	s := newTestService(t)

	rc := rpsl.RouterClass{
		"2001:db8::/32":      nil,
		"2001:db8:1::/48":    nil,
		"2001:db8:1:2::/64":  nil,
	}

	err := s.Build(context.Background(), rc)
	assert.NoError(t, err)

	// Should find all enclosing prefixes
	tags := s.FindTags(netip.MustParseAddr("2001:db8:1:2::1"))
	assert.Len(t, tags, 3)
	assert.Contains(t, tags, "2001:db8::/32")
	assert.Contains(t, tags, "2001:db8:1::/48")
	assert.Contains(t, tags, "2001:db8:1:2::/64")
}

func TestAtomicRebuild(t *testing.T) {
	s := newTestService(t)

	rc1 := rpsl.RouterClass{
		"2001:db8::/32": nil,
	}
	err := s.Build(context.Background(), rc1)
	assert.NoError(t, err)

	tag, found := s.FindDeepestTag(netip.MustParseAddr("2001:db8::1"))
	assert.True(t, found)
	assert.Equal(t, "2001:db8::/32", tag)

	// Rebuild with different data
	rc2 := rpsl.RouterClass{
		"2001:db9::/32": nil,
	}
	err = s.Build(context.Background(), rc2)
	assert.NoError(t, err)

	// Old prefix should be gone
	_, found = s.FindDeepestTag(netip.MustParseAddr("2001:db8::1"))
	assert.False(t, found)

	// New prefix should work
	tag, found = s.FindDeepestTag(netip.MustParseAddr("2001:db9::1"))
	assert.True(t, found)
	assert.Equal(t, "2001:db9::/32", tag)
}

func BenchmarkFindDeepestTag_IPv6(b *testing.B) {
	log := logger.NewSimple("bench")
	s := New(log)

	// Build a tree with realistic number of prefixes
	rc := make(rpsl.RouterClass)
	for i := 0; i < 1000; i++ {
		prefix := netip.MustParsePrefix("2001:db8::" + fmt.Sprintf("%x", i) + "/128")
		rc[prefix.String()] = nil
	}
	// Add some shorter prefixes
	rc["2001:db8::/32"] = nil
	rc["2001::/16"] = nil

	s.Build(context.Background(), rc)

	ip := netip.MustParseAddr("2001:db8::ff")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.FindDeepestTag(ip)
	}
}

func BenchmarkFindTags_IPv6(b *testing.B) {
	log := logger.NewSimple("bench")
	s := New(log)

	rc := make(rpsl.RouterClass)
	rc["2001:db8::/32"] = nil
	rc["2001:db8:1::/48"] = nil
	rc["2001:db8:1:2::/64"] = nil
	rc["2001::/16"] = nil

	s.Build(context.Background(), rc)

	ip := netip.MustParseAddr("2001:db8:1:2::1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.FindTags(ip)
	}
}
