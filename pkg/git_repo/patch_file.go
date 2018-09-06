package git_repo

import (
	"fmt"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
)

type PatchFile struct {
	FilePath string
}

func NewTmpPatchFile() *PatchFile {
	path := filepath.Join("/tmp", fmt.Sprintf("dapp-%s.patch", uuid.NewV4().String()))
	return &PatchFile{FilePath: path}
}

func (p *PatchFile) GetFilePath() string {
	return p.FilePath
}

func (p *PatchFile) IsEmpty() (bool, error) {
	fi, err := os.Stat(p.GetFilePath())
	if err != nil {
		return false, err
	}
	if fi.Size() > 0 {
		return false, nil
	}
	return true, nil
}
