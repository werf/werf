package utils

import (
	"archive/tar"
	"errors"
	"io"

	. "github.com/onsi/gomega"
)

// ForEachInTarballFunc returns true to continue, false to stop.
type ForEachInTarballFunc func(*tar.Header) error

func ForEachInTarball(reader *tar.Reader, handler ForEachInTarballFunc) error {
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			Expect(err).To(Succeed())
		}

		if err = handler(header); err != nil {
			return err
		}
	}

	return nil
}
