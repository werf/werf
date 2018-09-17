package git_repo

import (
	"fmt"
	"path/filepath"

	git_util "github.com/flant/dapp/pkg/git"
	uuid "github.com/satori/go.uuid"
)

type PatchFile struct {
	FilePath   string
	Descriptor *git_util.PatchDescriptor
}

func NewTmpPatchFile() *PatchFile {
	path := filepath.Join("/tmp", fmt.Sprintf("dapp-%s.patch", uuid.NewV4().String()))
	return &PatchFile{FilePath: path}
}

func (p *PatchFile) GetFilePath() string {
	return p.FilePath
}

func (p *PatchFile) IsEmpty() bool {
	return p.Descriptor.IsEmpty
}
