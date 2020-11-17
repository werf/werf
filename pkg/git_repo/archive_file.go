package git_repo

import (
	"github.com/werf/werf/pkg/true_git"
)

type ArchiveFile struct {
	FilePath   string
	Descriptor *true_git.ArchiveDescriptor
}

func (a *ArchiveFile) GetFilePath() string {
	return a.FilePath
}

func (a *ArchiveFile) GetType() ArchiveType {
	return ArchiveType(a.Descriptor.Type)
}

func (a *ArchiveFile) IsEmpty() bool {
	return a.Descriptor.IsEmpty
}
