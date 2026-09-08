package apiv1

import (
	"testing"

	"ip_service/pkg/contexthandler"

	"github.com/stretchr/testify/assert"
)

func TestIndex_OK(t *testing.T) {
	c := mockClient(t)
	ctx := contexthandler.Add(t.Context(), "request", &contexthandler.RequestContext{
		ClientIP:  "89.160.20.112",
		UserAgent: mockUserAgent,
	})

	got, err := c.Index(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "89.160.20.112", got.IP)
	assert.Equal(t, uint(29518), got.ASN)
	assert.Equal(t, "SE", got.CountryISO)
	assert.True(t, got.IsEU)
}

func TestAllJSON_OK(t *testing.T) {
	c := mockClient(t)
	ctx := contexthandler.Add(t.Context(), "request", &contexthandler.RequestContext{
		ClientIP:  "89.160.20.112",
		UserAgent: mockUserAgent,
	})

	got, err := c.AllJSON(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "Bredband2 AB", got.ASNOrganization)
}

func TestIPDecimal_OK(t *testing.T) {
	c := mockClient(t)
	ctx := contexthandler.Add(t.Context(), "request", &contexthandler.RequestContext{
		ClientIP: "127.0.0.1",
	})
	got, err := c.IPDecimal(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "2130706433", got)
}

func TestIPDecimal_MissingContext(t *testing.T) {
	c := mockClient(t)
	_, err := c.IPDecimal(t.Context())
	assert.Error(t, err)
}

func TestCityText_JSON(t *testing.T) {
	c := mockClient(t)
	ctx := contexthandler.Add(t.Context(), "request", &contexthandler.RequestContext{
		ClientIP: "89.160.20.112",
	})

	txt, err := c.CityText(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Linköping", txt)

	j, err := c.CityJSON(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Linköping", j["city"])
}

func TestCoordinates_TextJSON(t *testing.T) {
	c := mockClient(t)
	ctx := contexthandler.Add(t.Context(), "request", &contexthandler.RequestContext{
		ClientIP: "89.160.20.112",
	})

	coord, err := c.CoordinatesText(ctx)
	assert.NoError(t, err)
	assert.InDelta(t, 58.4167, coord.Latitude, 0.001)
	assert.InDelta(t, 15.6167, coord.Longitude, 0.001)

	j, err := c.CoordinatesJSON(ctx)
	assert.NoError(t, err)
	assert.Contains(t, j, "latitude")
	assert.Contains(t, j, "longitude")
}

func TestCollision(t *testing.T) {
	c := mockClient(t)

	got, err := c.Collision(t.Context(), &CollisionRequest{
		IP1: "10.0.0.0/8",
		IP2: "10.1.0.0/16",
	})
	assert.NoError(t, err)
	assert.True(t, got.Collision)

	got, err = c.Collision(t.Context(), &CollisionRequest{
		IP1: "10.0.0.0/8",
		IP2: "192.168.0.0/16",
	})
	assert.NoError(t, err)
	assert.False(t, got.Collision)

	_, err = c.Collision(t.Context(), &CollisionRequest{IP1: "bad", IP2: "10.0.0.0/8"})
	assert.Error(t, err)

	_, err = c.Collision(t.Context(), &CollisionRequest{IP1: "10.0.0.0/8", IP2: "bad"})
	assert.Error(t, err)
}

func TestGetIP_MissingContext(t *testing.T) {
	c := mockClient(t)
	_, err := c.getIP(t.Context())
	assert.Error(t, err)
}

func TestUA_MissingContext(t *testing.T) {
	c := mockClient(t)
	_, err := c.ua(t.Context())
	assert.Error(t, err)
}

func TestUA_Parses(t *testing.T) {
	c := mockClient(t)
	ctx := contexthandler.Add(t.Context(), "request", &contexthandler.RequestContext{
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64) Chrome/1",
	})
	parsed, err := c.ua(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed.Name)
}
