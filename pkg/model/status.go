package model

import (
	"fmt"
	"log"
	"time"
)

var (
	//StatusOK status ok
	StatusOK = "STATUS_OK_%s"
	// StatusFail status fail
	StatusFail = "STATUS_FAIL_%s"
)

// StatusProbeStore contains the previous probe result and the next time to check
type StatusProbeStore struct {
	NextCheck      time.Time
	PreviousResult *StatusProbe
}

// StatusProbes contains probes
type StatusProbes []*StatusProbe

var (
	// BuildVariableGitCommit contains ldflags -X variable git commit hash
	BuildVariableGitCommit string = "undef"

	// BuildVariableTimestamp contains ldflags -X variable build time
	BuildVariableTimestamp string = "undef"

	// BuildVariableGoVersion contains ldflags -X variable go build version
	BuildVariableGoVersion string = "undef"

	// BuildVariableGoArch contains ldflags -X variable go arch build
	BuildVariableGoArch string = "undef"

	// BuildVariableGitBranch contains ldflags -X variable git branch
	BuildVariableGitBranch string = "undef"

	// BuildVersion contains ldflags -X variable build version
	BuildVersion string = "undef"
)

// StatusProbe holds the status of a probe
type StatusProbe struct {
	Name          string         `json:"name,omitempty"`
	Healthy       bool           `json:"healthy,omitempty"`
	Message       map[string]any `json:"message,omitempty"`
	LastCheckedTS time.Time      `json:"timestamp,omitempty"`
}

type StatusReplyData struct {
	ServiceName    string          `json:"service_name,omitempty"`
	BuildVariables *BuildVariables `json:"build_variables,omitempty"`
	StatusProbes   StatusProbes    `json:"probes,omitempty"`
	Status         string          `json:"status,omitempty"`
}

type StatusReply struct {
	Data *StatusReplyData `json:"data,omitempty"`
}

type BuildVariables struct {
	GitCommit string `json:"git_commit,omitempty"`
	GitBranch string `json:"git_branch,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
	GoArch    string `json:"go_arch,omitempty"`
	Version   string `json:"version,omitempty"`
}

// Check checks the status of each status, return the first that does not pass.
func (probes StatusProbes) Check(serviceName string) *StatusReply {
	health := &StatusReply{
		Data: &StatusReplyData{
			ServiceName: serviceName,
			BuildVariables: &BuildVariables{
				GitCommit: BuildVariableGitCommit,
				GitBranch: BuildVariableGitBranch,
				Timestamp: BuildVariableTimestamp,
				GoVersion: BuildVariableGoVersion,
				GoArch:    BuildVariableGoArch,
				Version:   BuildVersion,
			},
			StatusProbes: StatusProbes{},
			Status:       fmt.Sprintf(StatusOK, serviceName),
		},
	}

	if probes == nil {
		log.Println("probe is nil")
		return health
	}

	for _, probe := range probes {
		if !probe.Healthy {
			health.Data.Status = fmt.Sprintf(StatusFail, serviceName)
		}
		health.Data.StatusProbes = append(health.Data.StatusProbes, probe)
	}

	return health
}
