package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/gomega"
)

var LineBreak = "\n"

func init() {
	if runtime.GOOS == "windows" {
		LineBreak = "\r\n"
	}
}

func CopyIn(sourcePath, destinationPath string) {
	err := os.MkdirAll(destinationPath, 0777)
	Ω(err).ShouldNot(HaveOccurred())

	files, err := ioutil.ReadDir(sourcePath)
	Ω(err).ShouldNot(HaveOccurred())

	copyFile := func(srcPath, dstPath string) {
		src, err := os.Open(srcPath)
		Ω(err).ShouldNot(HaveOccurred())
		defer func() { _ = src.Close() }()

		dst, err := os.Create(dstPath)
		Ω(err).ShouldNot(HaveOccurred())
		defer func() { _ = dst.Close() }()

		_, err = io.Copy(dst, src)
		Ω(err).ShouldNot(HaveOccurred())
	}

	for _, f := range files {
		srcPath := filepath.Join(sourcePath, f.Name())
		dstPath := filepath.Join(destinationPath, f.Name())
		if f.IsDir() {
			CopyIn(srcPath, dstPath)
			continue
		}

		copyFile(srcPath, dstPath)
	}
}

func CreateFile(path string, data []byte) {
	Ω(os.MkdirAll(filepath.Dir(path), 0777)).Should(Succeed())
	Ω(ioutil.WriteFile(path, data, 0644)).Should(Succeed())
}
