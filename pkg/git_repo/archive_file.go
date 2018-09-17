package git_repo

import (
	"fmt"
	"path/filepath"

	git_util "github.com/flant/dapp/pkg/git"
	uuid "github.com/satori/go.uuid"
)

type ArchiveFile struct {
	FilePath   string
	Descriptor *git_util.ArchiveDescriptor
}

func NewTmpArchiveFile() *ArchiveFile {
	path := filepath.Join("/tmp", fmt.Sprintf("dapp-%s.archive.tar", uuid.NewV4().String()))
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
