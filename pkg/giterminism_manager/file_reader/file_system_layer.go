package file_reader

import (
	"os"
	"path/filepath"

	"github.com/werf/common-go/pkg/util"
)

type fileSystemLayer struct{}

func newFileSystemOperator() fileSystemLayer {
	return fileSystemLayer{}
}

func (fso fileSystemLayer) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (fso fileSystemLayer) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

func (fso fileSystemLayer) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

func (fso fileSystemLayer) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func (fso fileSystemLayer) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (fso fileSystemLayer) FileExists(path string) (bool, error) {
	return util.FileExists(path)
}

func (fso fileSystemLayer) DirExists(path string) (bool, error) {
	return util.DirExists(path)
}

func (fso fileSystemLayer) RegularFileExists(path string) (bool, error) {
	return util.RegularFileExists(path)
}
