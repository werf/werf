//go:build !linux
// +build !linux

package thirdparty

import "time"

// Copied from github.com/containers/buildah@v1.26.1/docker/types.go:52
type BuildahHealthConfig struct {
	Test        []string      `json:",omitempty"`
	Interval    time.Duration `json:",omitempty"`
	Timeout     time.Duration `json:",omitempty"`
	StartPeriod time.Duration `json:",omitempty"`
	Retries     int           `json:",omitempty"`
}
