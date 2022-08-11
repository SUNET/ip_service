package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockStatusService1OK = &StatusService{
	Healthy:     true,
	Status:      StatusOK,
}

var mockStatusService2Fail = &StatusService{
	Healthy:     false,
	Status:      StatusFail,
}

func TestCheck(t *testing.T) {
	tts := []struct {
		name string
		have AllStatus
		want string
	}{
		{
			name: "OK",
			have: []*StatusService{
				mockStatusService1OK,
			},
			want: StatusOK,
		},
		{
			name: "Fail",
			have: []*StatusService{
				mockStatusService1OK,
				mockStatusService2Fail,
			},
			want: StatusFail,
		},
		{
			name: "Fail again",
			have: []*StatusService{
				mockStatusService2Fail,
				mockStatusService1OK,
			},
			want: StatusFail,
		},
		{
			name: "nil",
			have: nil,
			want: StatusFail,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.have.Check()
			assert.Equal(t, tt.want, got.Status)
		})
	}
}
