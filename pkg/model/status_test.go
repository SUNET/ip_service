package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusProbes_Check_Nil(t *testing.T) {
	var probes StatusProbes
	got := probes.Check("svc")
	assert.NotNil(t, got)
	assert.Equal(t, "svc", got.Data.ServiceName)
	assert.Equal(t, "STATUS_OK_svc", got.Data.Status)
	assert.Empty(t, got.Data.StatusProbes)
}

func TestStatusProbes_Check_AllHealthy(t *testing.T) {
	probes := StatusProbes{
		&StatusProbe{Name: "a", Healthy: true},
		&StatusProbe{Name: "b", Healthy: true},
	}
	got := probes.Check("svc")
	assert.Equal(t, "STATUS_OK_svc", got.Data.Status)
	assert.Len(t, got.Data.StatusProbes, 2)
}

func TestStatusProbes_Check_OneFailing(t *testing.T) {
	probes := StatusProbes{
		&StatusProbe{Name: "a", Healthy: true},
		&StatusProbe{Name: "b", Healthy: false},
	}
	got := probes.Check("svc")
	assert.Equal(t, "STATUS_FAIL_svc", got.Data.Status)
	assert.Len(t, got.Data.StatusProbes, 2)
}
