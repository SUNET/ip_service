package lctree

import (
	"testing"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestClose(t *testing.T) {
	s := New(logger.NewSimple("test"))
	assert.NoError(t, s.Close(t.Context()))
}
