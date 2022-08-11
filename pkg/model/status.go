package model

import (
	"time"
)

var (
	// StatusOK ok
	StatusOK = "STATUS_OK_ip-service_"
	// StatusFail not ok
	StatusFail = "STATUS_FAIL_ip-service_"
)

// StatusService type
type StatusService struct {
	ServiceName string        `json:"service_name,omitempty"`
	Message     string        `json:"message,omitempty"`
	Healthy     bool          `json:"healthy,omitempty"`
	Status      string        `json:"status,omitempty"`
	Timestamp   time.Time     `json:"timestamp,omitempty"`
	Interval    time.Duration `json:"-"`
}

type AllStatus []*StatusService

// Check checks the status of each status, return the first that does not pass.
func (s AllStatus) Check() *StatusService {
	if s == nil {
		return &StatusService{
			Healthy:   false,
			Status:    StatusFail,
			Timestamp: time.Now(),
		}
	}

	for _, status := range s {
		if !status.Healthy {
			status.Status = StatusFail
			return status
		}
	}
	status := &StatusService{
		Healthy:   true,
		Status:    StatusOK,
		Timestamp: time.Now(),
	}
	return status
}
