package file_reader

import (
	"github.com/werf/werf/pkg/giterminism"
)

type FileReader struct {
	manager giterminism.Manager
}

func NewFileReader(manager giterminism.Manager) FileReader {
	return FileReader{manager}
}
