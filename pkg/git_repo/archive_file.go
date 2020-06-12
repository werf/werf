package git_repo

import (
	"fmt"
	"path/filepath"

	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
	uuid "github.com/satori/go.uuid"
)

type ArchiveFile struct {
	FilePath   string
	Descriptor *true_git.ArchiveDescriptor
}

func NewTmpArchiveFile() *ArchiveFile {
	path := filepath.Join(werf.GetTmpDir(), fmt.Sprintf("werf-%s.archive.tar", uuid.NewV4().String()))
	return &ArchiveFile{FilePath: path}
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
