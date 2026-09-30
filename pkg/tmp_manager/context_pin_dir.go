package tmp_manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// CreateContextPinDir creates a directory under $WERF_HOME, on the same filesystem as the git
// archive cache, so that a cached archive can be hard-linked into it instead of copied.
func CreateContextPinDir(ctx context.Context) (string, error) {
	pinsDir := filepath.Join(getServiceTmpDir(), contextPinsServiceDir)
	if err := os.MkdirAll(pinsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("create context pins dir %s: %w", pinsDir, err)
	}

	newDir, err := os.MkdirTemp(pinsDir, contextPinDirPrefix)
	if err != nil {
		return "", fmt.Errorf("create context pin dir: %w", err)
	}

	if err := registrator.queueRegistration(ctx, newDir, filepath.Join(getCreatedTmpDirs(), contextPinsServiceDir)); err != nil {
		return "", fmt.Errorf("unable to queue GC registration: %w", err)
	}

	return newDir, nil
}
