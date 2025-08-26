package image

import (
	"context"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// ImageSpecConfig represents OCI image configuration
// https://github.com/opencontainers/image-spec/blob/main/config.md
type SpecConfig struct {
	Created         string              `json:"created"`
	Author          string              `json:"author"`
	User            string              `json:"User"`
	ExposedPorts    map[string]struct{} `json:"ExposedPorts"`
	Env             []string            `json:"Env"`
	Entrypoint      []string            `json:"Entrypoint"`
	Cmd             []string            `json:"Cmd"`
	Volumes         map[string]struct{} `json:"Volumes"`
	WorkingDir      string              `json:"WorkingDir"`
	Labels          map[string]string   `json:"Labels"`
	StopSignal      string              `json:"StopSignal"`
	HealthConfig    *HealthConfig       `json:"Healthcheck,omitempty"`
	ClearHistory    bool
	ClearCmd        bool
	ClearEntrypoint bool
	ClearUser       bool
	ClearWorkingDir bool
}

type HealthConfig struct {
	Test        []string      `json:",omitempty"`
	Interval    time.Duration `json:",omitempty"`
	Timeout     time.Duration `json:",omitempty"`
	StartPeriod time.Duration `json:",omitempty"`
	Retries     int           `json:",omitempty"`
}

type Storage interface {
	MutateAndPushImage(ctx context.Context, src, dest string, newConfig SpecConfig) error
}

func MutateAndPushImage(ctx context.Context, storage Storage, sourceReference, destinationReference string, newConfig SpecConfig) error {
	return storage.MutateAndPushImage(ctx, sourceReference, destinationReference, newConfig)
}

func UpdateConfigFile(updates SpecConfig, target *v1.ConfigFile) {
	if updates.Author != "" {
		target.Author = updates.Author
	}
	if updates.ClearHistory {
		target.History = []v1.History{}
	}
	if updates.Volumes != nil {
		target.Config.Volumes = updates.Volumes
	}
	if updates.Labels != nil {
		target.Config.Labels = updates.Labels
	}

	target.Config.Env = updates.Env

	if updates.ExposedPorts != nil {
		target.Config.ExposedPorts = updates.ExposedPorts
	}
	if updates.ClearUser {
		target.Config.User = ""
	}
	if updates.User != "" {
		target.Config.User = updates.User
	}
	if updates.ClearCmd {
		target.Config.Cmd = []string{}
	}
	if len(updates.Cmd) > 0 {
		target.Config.Cmd = updates.Cmd
	}
	if updates.ClearEntrypoint {
		target.Config.Entrypoint = []string{}
	}
	if len(updates.Entrypoint) > 0 {
		target.Config.Entrypoint = updates.Entrypoint
	}
	if updates.ClearWorkingDir {
		target.Config.WorkingDir = ""
	}
	if updates.WorkingDir != "" {
		target.Config.WorkingDir = updates.WorkingDir
	}
	if updates.StopSignal != "" {
		target.Config.StopSignal = updates.StopSignal
	}
	if updates.HealthConfig != nil {
		target.Config.Healthcheck = &v1.HealthConfig{
			Test:        updates.HealthConfig.Test,
			Interval:    updates.HealthConfig.Interval,
			Timeout:     updates.HealthConfig.Timeout,
			StartPeriod: updates.HealthConfig.StartPeriod,
			Retries:     updates.HealthConfig.Retries,
		}
	}
}
