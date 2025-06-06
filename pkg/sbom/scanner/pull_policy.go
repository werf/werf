package scanner

import (
	"github.com/containers/buildah/define"
)

type PullPolicy define.PullPolicy

const (
	// PullIfMissing scanner image should be pulled from a registry if a local copy of it is not already present.
	PullIfMissing = PullPolicy(define.PullIfMissing)
	// PullAlways scanner image should be pulled from a registry before the build proceeds.
	PullAlways = PullPolicy(define.PullAlways)
	// PullIfNewer scanner image should only be pulled from a registry if a local copy is not already present or if a
	// newer version the image is present on the repository.
	// PullIfNewer = PullPolicy(define.PullIfNewer) NOTE: docker does not support this option value
	// PullNever scanner image should not be pulled from a registry.
	PullNever = PullPolicy(define.PullNever)
)

func (p PullPolicy) String() string {
	return define.PullPolicy(p).String()
}
