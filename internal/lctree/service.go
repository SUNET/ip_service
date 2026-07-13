package lctree

import (
	"context"
	"github.com/SUNET/vc/pkg/logger"
	"ip_service/pkg/rpsl"
	"net/netip"
	"sync"

	patricia "github.com/kentik/patricia"
	tree "github.com/kentik/patricia/generics_tree"
)

// Service wraps two patricia trees (IPv4 + IPv6) for fast longest-prefix-match lookups.
// Tags are network prefix strings (e.g. "2001:db8::/32") matching rpsl.RouterClass keys.
type Service struct {
	v4  *tree.TreeV4[string]
	v6  *tree.TreeV6[string]
	mu  sync.RWMutex
	log *logger.Log
}

func New(log *logger.Log) *Service {
	log.Info("Started")
	return &Service{
		v4:  tree.NewTreeV4[string](),
		v6:  tree.NewTreeV6[string](),
		log: log,
	}
}

// Build populates the trees from a RouterClass map. It builds new trees and swaps them in atomically.
func (s *Service) Build(ctx context.Context, routerClass rpsl.RouterClass) error {
	v4 := tree.NewTreeV4[string]()
	v6 := tree.NewTreeV6[string]()

	var v4Count, v6Count int

	for network := range routerClass {
		prefix, err := netip.ParsePrefix(network)
		if err != nil {
			s.log.Debug("skipping invalid prefix", "network", network, "error", err)
			continue
		}

		if prefix.Addr().Is4() {
			v4Addr, _, err := patricia.ParseFromNetIPPrefix(prefix)
			if err != nil {
				s.log.Debug("skipping v4 prefix parse error", "network", network, "error", err)
				continue
			}
			v4.Add(*v4Addr, network, nil)
			v4Count++
		} else {
			_, v6Addr, err := patricia.ParseFromNetIPPrefix(prefix)
			if err != nil {
				s.log.Debug("skipping v6 prefix parse error", "network", network, "error", err)
				continue
			}
			v6.Add(*v6Addr, network, nil)
			v6Count++
		}
	}

	s.mu.Lock()
	s.v4 = v4
	s.v6 = v6
	s.mu.Unlock()

	s.log.Info("Patricia tree built", "v4_prefixes", v4Count, "v6_prefixes", v6Count)
	return nil
}

// FindTags returns all network prefixes that contain the given IP (from least to most specific).
func (s *Service) FindTags(ip netip.Addr) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ip.Is4() {
		v4Addr, _, err := patricia.ParseFromNetIPAddr(ip)
		if err != nil {
			return nil
		}
		return s.v4.FindTags(*v4Addr)
	}

	_, v6Addr, err := patricia.ParseFromNetIPAddr(ip)
	if err != nil {
		return nil
	}
	return s.v6.FindTags(*v6Addr)
}

// FindDeepestTag returns the most-specific (longest prefix) network containing the IP.
func (s *Service) FindDeepestTag(ip netip.Addr) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ip.Is4() {
		v4Addr, _, err := patricia.ParseFromNetIPAddr(ip)
		if err != nil {
			return "", false
		}
		found, tag := s.v4.FindDeepestTag(*v4Addr)
		return tag, found
	}

	_, v6Addr, err := patricia.ParseFromNetIPAddr(ip)
	if err != nil {
		return "", false
	}
	found, tag := s.v6.FindDeepestTag(*v6Addr)
	return tag, found
}

// CountTags returns the total number of tags across both trees.
func (s *Service) CountTags() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.v4.CountTags(), s.v6.CountTags()
}
