package tmp_manager

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
)

func CreateDockerConfigDir(ctx context.Context, fromDockerConfig string) (string, error) {
	newDir, err := newTmpDir(dockerConfigDirPrefix)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(newDir, 0o700); err != nil {
		return "", err
	}

	if _, err := os.Stat(fromDockerConfig); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(fromDockerConfig)
		if err != nil {
			return "", fmt.Errorf("unable to read %q", fromDockerConfig)
		}

		pathsToSkip := []string{"run", "mutagen"}
	mainLoop:
		for _, file := range files {
			for _, pathToSkip := range pathsToSkip {
				if file.Name() == pathToSkip {
					continue mainLoop
				}
			}

			source := filepath.Join(fromDockerConfig, file.Name())
			destination := filepath.Join(newDir, file.Name())
			err := copy.Copy(source, destination)
			if err != nil {
				return "", fmt.Errorf("unable to copy %q to %q: %w", source, destination, err)
			}
		}
	}

	if err = registrator.queueRegistration(ctx, newDir, filepath.Join(getCreatedTmpDirs(), dockerConfigsServiceDir)); err != nil {
		return "", fmt.Errorf("unable to queue GC registration: %w", err)
	}

	return newDir, nil
}
