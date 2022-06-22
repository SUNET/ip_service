package store

import (
	"context"
	"ip_service/pkg/logger"
	"ip_service/pkg/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockNew(t *testing.T) *Service {
	cfg := &model.Cfg{}

	s, err := New(context.TODO(), cfg, logger.New("test", false).New("test"))
	assert.NoError(t, err)

	return s
}
