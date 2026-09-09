package apiv1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests exercise error paths where the request context is missing.
func TestMethods_MissingContext(t *testing.T) {
	c := mockClient(t)

	_, err := c.asn(t.Context())
	assert.Error(t, err)

	_, err = c.city(t.Context())
	assert.Error(t, err)

	_, err = c.country(t.Context())
	assert.Error(t, err)

	_, err = c.countryISO(t.Context())
	assert.Error(t, err)

	_, err = c.coordinates(t.Context())
	assert.Error(t, err)

	_, err = c.getUserAgent(t.Context())
	assert.Error(t, err)

	_, err = c.formatAllJSON(t.Context())
	assert.Error(t, err)

	_, err = c.formatLookUpJSON(t.Context())
	assert.Error(t, err)
}
