package extract

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// FromImageStream extracts the contents of an image stream (tarball) into a flattened filesystem tarball.
func FromImageStream(opener tarball.Opener) (io.ReadCloser, error) {
	if img, err := tarball.Image(opener, nil); err != nil {
		return nil, fmt.Errorf("unable to open image: %w", err)
	} else {
		return mutate.Extract(img), nil
	}
}

// FromImageBytes extracts the first SBOM artifact in the tar stream.
// It assumes that the tar stream contains only one artifact file.
func FromImageBytes(opener tarball.Opener) (data []byte, errOut error) {
	readerCloser, err := FromImageStream(opener)
	if err != nil {
		return nil, fmt.Errorf("unable to extract SBOM from tar: %w", err)
	}
	defer func() {
		if err = readerCloser.Close(); err != nil {
			errOut = errors.Join(err, fmt.Errorf("unable to close tar reader: %w", err))
		}
	}()

	reader := tar.NewReader(readerCloser)

	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unable to read tar header: %w", err)
		}

		if header.FileInfo().IsDir() {
			continue
		}

		// TODO: assume we have only one artifact file
		data = make([]byte, header.Size)
		if _, err = io.ReadFull(reader, data); err != nil {
			return nil, fmt.Errorf("unable to read artifact file content: %w", err)
		}
		return data, nil
	}

	return nil, errors.New("no artifact file found")
}
