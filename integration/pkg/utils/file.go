package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/otiai10/copy"

	. "github.com/onsi/gomega"
)

var LineBreak = "\n"

func init() {
	if runtime.GOOS == "windows" {
		LineBreak = "\r\n"
	}
}

func CopyIn(sourcePath, destinationPath string) {
	Ω(copy.Copy(sourcePath, destinationPath)).Should(Succeed())
}

func MkdirAll(dir string) {
	Ω(os.MkdirAll(dir, 0o777)).Should(Succeed())
}

func WriteFile(path string, data []byte) {
	Ω(os.MkdirAll(filepath.Dir(path), 0o777)).Should(Succeed())
	Ω(ioutil.WriteFile(path, data, 0o644)).Should(Succeed())
}

func DeleteFile(path string) {
	Ω(os.Remove(path)).Should(Succeed())
}
