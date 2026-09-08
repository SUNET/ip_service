package maxmind

import (
	"path/filepath"
	"testing"

	"ip_service/internal/store"
	"ip_service/pkg/model"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/SUNET/vc/pkg/trace"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
)

func newServiceWithTestDBs(t *testing.T) *Service {
	t.Helper()
	ctx := t.Context()

	dbCity, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-city-Test.mmdb"))
	assert.NoError(t, err)
	dbASN, err := geoip2.Open(filepath.Join("..", "..", "testdata", "GeoLite2-asn-Test.mmdb"))
	assert.NoError(t, err)

	tp, err := trace.NewForTesting(ctx, "test", logger.NewSimple("test"))
	assert.NoError(t, err)

	cfg := &model.Cfg{IPService: &model.IPService{
		Store: model.Store{File: model.FileStorage{Path: t.TempDir()}},
	}}
	st, err := store.New(ctx, cfg, tp, logger.NewSimple("store"))
	assert.NoError(t, err)

	return &Service{
		probeStore: &model.StatusProbeStore{},
		cfg:        cfg,
		Log:        logger.NewSimple("maxmind"),
		TP:         tp,
		kvStore:    st.KV,
		DBASN:      dbASN,
		DBCity:     dbCity,
		DBMeta:     DBMeta{model.MaxmindDBTypeASN: &DBObject{}, model.MaxmindDBTypeCity: &DBObject{}},
	}
}

func TestStatus_HealthyBothDBs(t *testing.T) {
	s := newServiceWithTestDBs(t)

	probe := s.Status(t.Context())
	assert.NotNil(t, probe)
	assert.Equal(t, "maxmind", probe.Name)
	assert.True(t, probe.Healthy)

	// Cached on subsequent call.
	probe2 := s.Status(t.Context())
	assert.Same(t, probe, probe2)
}

func TestStatus_MissingDBs(t *testing.T) {
	s := newServiceWithTestDBs(t)
	s.DBASN = nil
	s.DBCity = nil

	probe := s.Status(t.Context())
	assert.NotNil(t, probe)
	assert.False(t, probe.Healthy)
	assert.Equal(t, "unavailable", probe.Message["asn_db_status"])
	assert.Equal(t, "unavailable", probe.Message["city_db_status"])
}

func TestASN_City_ISP_Anonymous(t *testing.T) {
	s := newServiceWithTestDBs(t)
	ctx := t.Context()

	ip := netParse("89.160.20.112")
	asn, err := s.ASN(ctx, ip)
	assert.NoError(t, err)
	assert.Equal(t, uint(29518), asn.AutonomousSystemNumber)

	city, err := s.City(ctx, ip)
	assert.NoError(t, err)
	assert.Equal(t, "Link\u00f6ping", city.City.Names["en"])

	// ISP/AnonymousIP use the DB.Reader methods; the test DB may or may not
	// support these but they should not panic.
	_, _ = s.ISP(ctx, ip)
	_, _ = s.AnonymousIP(ctx, ip)
}

func TestASN_NilDB(t *testing.T) {
	s := newServiceWithTestDBs(t)
	s.DBASN = nil
	_, err := s.ASN(t.Context(), netParse("1.2.3.4"))
	assert.Error(t, err)
}

func TestCity_NilDB(t *testing.T) {
	s := newServiceWithTestDBs(t)
	s.DBCity = nil
	_, err := s.City(t.Context(), netParse("1.2.3.4"))
	assert.Error(t, err)
}
