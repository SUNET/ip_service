package rpslsource

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// closedPortHost returns a "host:port" that will refuse connections almost immediately.
func closedPortHost(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	addr := l.Addr().String()
	// Close it so future dials fail with connection refused.
	assert.NoError(t, l.Close())
	return addr
}

func TestGetRemoteSerialFTP_DialError(t *testing.T) {
	host := closedPortHost(t)

	s := mockService(t, Config{
		Name:      "radb",
		Transport: TransportFTP,
		RemoteFiles: []RemoteFile{
			{Name: "radb", Path: "/radb/dbase/radb.db.gz"},
		},
		SerialPath: "/radb/dbase/RADB.CURRENTSERIAL",
		Host:       host,
	})

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	_, err := s.getRemoteSerialFTP(ctx)
	assert.Error(t, err)
}

func TestDownloadArchiveFTP_DialError(t *testing.T) {
	host := closedPortHost(t)

	s := mockService(t, Config{
		Name:      "radb",
		Transport: TransportFTP,
		RemoteFiles: []RemoteFile{
			{Name: "radb", Path: "/radb/dbase/radb.db.gz"},
		},
		SerialPath: "/radb/dbase/RADB.CURRENTSERIAL",
		Host:       host,
	})

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	err := s.downloadArchiveFTP(ctx, RemoteFile{Name: "radb", Path: "/radb/dbase/radb.db.gz"})
	assert.Error(t, err)
}
