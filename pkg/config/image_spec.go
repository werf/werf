package config

import "slices"

type ImageSpec struct {
	Author           string            `yaml:"author,omitempty"`
	ClearHistory    bool     `yaml:"clearHistory,omitempty"`
	ClearWerfLabels bool     `yaml:"clearWerfLabels,omitempty"`
	RemoveLabels    []string `yaml:"removeLabels,omitempty"`
	RemoveVolumes    []string          `yaml:"removeVolumes,omitempty"`
	RemoveEnv        []string          `yaml:"removeEnv,omitempty"`
	ClearCmd         bool              `yaml:"clearCmd,omitempty"`
	ClearEntrypoint  bool              `yaml:"clearEntrypoint,omitempty"`
	Volumes          []string          `yaml:"volumes,omitempty"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	Env              map[string]string `yaml:"env,omitempty"`
	User             string            `yaml:"user,omitempty"`
	Cmd              []string          `yaml:"cmd,omitempty"`
	Entrypoint       []string          `yaml:"entrypoint,omitempty"`
	WorkingDir       string            `yaml:"workingDir,omitempty"`
	StopSignal       string            `yaml:"stopSignal,omitempty"`
	Expose           []string          `yaml:"expose,omitempty"`
	Healthcheck      *healthConfig     `yaml:"healthcheck,omitempty"`

	raw       *rawImageSpec
	rawGlobal *rawImageSpecGlobal
}

func mergeImageSpec(priority, fallback *ImageSpec) ImageSpec {
	if priority == nil {
		priority = &ImageSpec{}
	}

	var merged ImageSpec
	if fallback == nil {
		fallback = &ImageSpec{}
	} else {
		merged = *fallback
	}

	if priority.Author != "" {
		merged.Author = priority.Author
	}
	if priority.ClearHistory {
		merged.ClearHistory = priority.ClearHistory
	}
	if priority.ClearWerfLabels {
		merged.ClearWerfLabels = priority.ClearWerfLabels
	}

	merged.RemoveLabels = mergeSlices(fallback.RemoveLabels, priority.RemoveLabels)
	merged.RemoveVolumes = mergeSlices(fallback.RemoveVolumes, priority.RemoveVolumes)
	merged.RemoveEnv = mergeSlices(fallback.RemoveEnv, priority.RemoveEnv)
	merged.Volumes = mergeSlices(fallback.Volumes, priority.Volumes)

	if priority.Labels != nil {
		if merged.Labels == nil {
			merged.Labels = make(map[string]string)
		}
		for k, v := range priority.Labels {
			merged.Labels[k] = v
		}
	}
	if priority.Env != nil {
		if merged.Env == nil {
			merged.Env = make(map[string]string)
		}
		for k, v := range priority.Env {
			merged.Env[k] = v
		}
	}
	if priority.User != "" {
		merged.User = priority.User
	}
	if len(priority.Cmd) > 0 {
		merged.Cmd = priority.Cmd
	}
	if len(priority.Entrypoint) > 0 {
		merged.Entrypoint = priority.Entrypoint
	}
	if priority.WorkingDir != "" {
		merged.WorkingDir = priority.WorkingDir
	}
	if priority.StopSignal != "" {
		merged.StopSignal = priority.StopSignal
	}
	if len(priority.Expose) > 0 {
		merged.Expose = priority.Expose
	}
	if priority.Healthcheck != nil {
		merged.Healthcheck = priority.Healthcheck
	}

	return merged
}

func mergeSlices(a, b []string) []string {
	merged := append([]string{}, a...)
	for _, item := range b {
		if !slices.Contains(merged, item) {
			merged = append(merged, item)
		}
	}
	return merged
}
