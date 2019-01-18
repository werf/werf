package git_repo

import (
	"fmt"
	"path/filepath"

	"github.com/flant/werf/pkg/true_git"
	uuid "github.com/satori/go.uuid"
)

type PatchFile struct {
	FilePath   string
	Descriptor *true_git.PatchDescriptor
}

func NewTmpPatchFile() *PatchFile {
	path := filepath.Join("/tmp", fmt.Sprintf("werf-%s.patch", uuid.NewV4().String()))
	return &PatchFile{FilePath: path}
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
