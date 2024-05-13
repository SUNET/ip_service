package maxmind

import (
	"context"
	"fmt"
	"ip_service/pkg/model"
	"net"
	"time"
)

// Status return status for maxmind database and last saved database version
func (s *Service) Status(ctx context.Context) *model.StatusProbe {
	if time.Now().Before(s.probeStore.NextCheck) {
		return s.probeStore.PreviousResult
	}

	probe := &model.StatusProbe{
		Name:    "maxmind",
		Healthy: true,
		Message: map[string]any{
			"asn_db_status":   "ok",
			"city_db_status":  "ok",
			"asn_db_version":  "n/a",
			"asn_last_check":  "n/a",
			"city_db_version": "n/a",
			"city_last_check": "n/a",
		},
		LastCheckedTS: time.Now(),
	}

	for _, testIP := range []string{"95.142.107.181", "110.50.243.6", "69.162.81.155"} {
		_, err := s.DBASN.ASN(net.ParseIP(testIP))
		if err != nil {
			probe.Message["asn"] = fmt.Sprintf("%v", err)
			probe.Healthy = false
		}
		_, err = s.DBCity.Country(net.ParseIP(testIP))
		if err != nil {
			probe.Message["city"] = fmt.Sprintf("%v", err)
			probe.Healthy = false
		}
	}

	remoteVersionASN := s.kvStore.GetRemoteVersion(ctx, "asn")
	if remoteVersionASN != "" {
		probe.Message["asn_db_version"] = remoteVersionASN
	}
	remoteVersionCity := s.kvStore.GetRemoteVersion(ctx, "city")
	if remoteVersionCity != "" {
		probe.Message["city_db_version"] = remoteVersionCity
	}

	lastCheckASN := s.kvStore.GetLastChecked(ctx, "asn")
	if lastCheckASN != "" {
		probe.Message["asn_last_check"] = lastCheckASN
	}
	lastCheckCity := s.kvStore.GetLastChecked(ctx, "city")
	if lastCheckCity != "" {
		probe.Message["city_last_check"] = lastCheckCity
	}

	s.probeStore.PreviousResult = probe
	s.probeStore.NextCheck = time.Now().Add(10 * time.Second)

	return probe
}
