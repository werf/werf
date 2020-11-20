package tmp_manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
)

func GetContextTmpDir() string {
	return filepath.Join(GetServiceTmpDir(), "context")
}

func CreateTmpContextArchive(ctx context.Context, contextArchivePath string) (string, error) {
	path := filepath.Join(GetContextTmpDir(), uuid.NewV4().String())
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return "", fmt.Errorf("unable to create dir %q: %s", filepath.Dir(path), err)
	}

	source, err := os.Open(contextArchivePath)
	if err != nil {
		return "", fmt.Errorf("unable to open %q: %s", contextArchivePath, err)
	}
	defer source.Close()

	destination, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("unable to create %q: %s", path, err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return "", fmt.Errorf("error copying %q to %q: %s", contextArchivePath, path, err)
	}

	return path, nil
}
