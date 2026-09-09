package maxmind

import (
	"net"

	"github.com/SUNET/vc/pkg/logger"
)

// netParse is a small helper for tests to parse an IPv4/IPv6 string.
func netParse(s string) net.IP {
	return net.ParseIP(s)
}

func mustSimpleLog() *logger.Log {
	return logger.NewSimple("maxmind-test")
}
