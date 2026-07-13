package whois

import (
	"context"
	"ip_service/pkg/rpsl"
	"net/netip"
)

type QueryIPResponse map[string]*rpsl.Object

// QueryIP returns the rpsl ASN map for the most specific prefix containing the given IP.
func (s *Service) QueryIP(ctx context.Context, ip string) (rpsl.ASN, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, err := netip.ParseAddr(ip)
	if err != nil {
		return nil, err
	}

	network, found := s.tree.FindDeepestTag(n)
	if !found {
		return nil, nil
	}

	return s.RPSLRouterClass[network], nil
}

// QueryIPAll returns rpsl ASN maps for all matching prefixes (from least to most specific).
func (s *Service) QueryIPAll(ctx context.Context, ip netip.Addr) ([]rpsl.ASN, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	networks := s.tree.FindTags(ip)
	if len(networks) == 0 {
		return nil, nil
	}

	reply := make([]rpsl.ASN, 0, len(networks))
	for _, network := range networks {
		if asn, ok := s.RPSLRouterClass[network]; ok {
			reply = append(reply, asn)
		}
	}

	return reply, nil
}
