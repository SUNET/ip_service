package contexthandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAndGet(t *testing.T) {
	rc := &RequestContext{
		ClientIP:  "10.0.0.1",
		LookupIP:  "10.0.0.2",
		UserAgent: "ua/1.0",
		Accept:    "application/json",
	}

	ctx := Add(t.Context(), "request", rc)
	got, err := Get(ctx, "request")
	assert.NoError(t, err)
	assert.Same(t, rc, got)
}

func TestGet_MissingKey(t *testing.T) {
	_, err := Get(t.Context(), "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no value found")
}
