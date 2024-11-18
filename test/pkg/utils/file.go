package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
)

var LineBreak = "\n"

func init() {
	if runtime.GOOS == "windows" {
		LineBreak = "\r\n"
	}
}

func CopyIn(sourcePath, destinationPath string) {
	Expect(copy.Copy(sourcePath, destinationPath)).Should(Succeed())
}

func MkdirAll(dir string) {
	Expect(os.MkdirAll(dir, 0o777)).Should(Succeed())
}

func WriteFile(path string, data []byte) {
	Expect(os.MkdirAll(filepath.Dir(path), 0o777)).Should(Succeed())
	Expect(ioutil.WriteFile(path, data, 0o644)).Should(Succeed())
}

func DeleteFile(path string) {
	Expect(os.Remove(path)).Should(Succeed())
}
