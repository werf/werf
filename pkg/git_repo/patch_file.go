package git_repo

import (
	"github.com/werf/werf/pkg/true_git"
)

type PatchFile struct {
	FilePath   string
	Descriptor *true_git.PatchDescriptor
}

func (p *PatchFile) GetFilePath() string {
	return p.FilePath
}

func (p *PatchFile) IsEmpty() bool {
	return len(p.Descriptor.Paths) == 0
}

func (p *PatchFile) HasBinary() bool {
	return len(p.Descriptor.BinaryPaths) > 0
}

func (p *PatchFile) GetPaths() []string {
	return p.Descriptor.Paths
}

func (p *PatchFile) GetBinaryPaths() []string {
	return p.Descriptor.BinaryPaths
}

func (p *PatchFile) GetPathsToRemove() []string {
	return p.Descriptor.PathsToRemove
}
