package config

import (
	"slices"
)

type ImageSpec struct {
	Author                  string            `yaml:"author,omitempty"`
	ClearHistory            bool              `yaml:"clearHistory,omitempty"`
	KeepEssentialWerfLabels bool              `yaml:"keepEssentialWerfLabels,omitempty"`
	ClearWerfLabels         bool              `yaml:"clearWerfLabels,omitempty"` // TODO: remove in v3.
	RemoveLabels            []string          `yaml:"removeLabels,omitempty"`
	RemoveVolumes           []string          `yaml:"removeVolumes,omitempty"`
	RemoveEnv               []string          `yaml:"removeEnv,omitempty"`
	ClearCmd                bool              `yaml:"clearCmd,omitempty"`
	ClearEntrypoint         bool              `yaml:"clearEntrypoint,omitempty"`
	ClearUser               bool              `yaml:"clearUser,omitempty"`
	ClearWorkingDir         bool              `yaml:"clearWorkingDir,omitempty"`
	Volumes                 []string          `yaml:"volumes,omitempty"`
	Labels                  map[string]string `yaml:"labels,omitempty"`
	Env                     map[string]string `yaml:"env,omitempty"`
	User                    string            `yaml:"user,omitempty"`
	Cmd                     []string          `yaml:"cmd,omitempty"`
	Entrypoint              []string          `yaml:"entrypoint,omitempty"`
	WorkingDir              string            `yaml:"workingDir,omitempty"`
	StopSignal              string            `yaml:"stopSignal,omitempty"`
	Expose                  []string          `yaml:"expose,omitempty"`
	Healthcheck             *healthConfig     `yaml:"healthcheck,omitempty"`

	raw       *rawImageSpec
	rawGlobal *rawImageSpecGlobal
}

func mergeImageSpec(meta, image *ImageSpec) ImageSpec {
	if image == nil {
		image = &ImageSpec{}
	}

	if meta.Author != "" {
		image.Author = meta.Author
	}

	if meta.ClearHistory {
		image.ClearHistory = true
	}

	if meta.KeepEssentialWerfLabels {
		image.KeepEssentialWerfLabels = true
	}

	if meta.Labels != nil {
		if image.Labels == nil {
			image.Labels = meta.Labels
		} else {
			for k, v := range meta.Labels {
				image.Labels[k] = v
			}
		}
	}

	image.RemoveLabels = mergeSlices(image.RemoveLabels, meta.RemoveLabels)

	return *image

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
