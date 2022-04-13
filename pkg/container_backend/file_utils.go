package container_backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/logboek"
)

func removeExactPath(ctx context.Context, path string) error {
	_, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
	case err != nil:
		return fmt.Errorf("unable to access path %q: %w", path, err)
	default:
		logboek.Context(ctx).Debug().LogF("Removing path %s\n", path)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("unable to remove path %s: %w", path, err)
		}
	}

	return nil
}

func removeExactPathWithEmptyParentDirs(ctx context.Context, path string, keepParentDirs []string) error {
	_, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
	case err != nil:
		return fmt.Errorf("unable to access path %q: %w", path, err)
	default:
		logboek.Context(ctx).Debug().LogF("Removing path %s\n", path)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("unable to remove path %s: %w", path, err)
		}
	}

	p := path
	for {
		parentDir := filepath.Dir(p)
		if parentDir == p {
			return nil
		}
		p = parentDir

		for _, keepPath := range keepParentDirs {
			if keepPath == p {
				return nil
			}
		}

		_, err := os.Stat(p)
		switch {
		case os.IsNotExist(err):
			// This may happen when initially given input path is not exists
			continue
		case err != nil:
			return fmt.Errorf("unable to access path %q: %w", path, err)
		}

		entries, err := os.ReadDir(p)
		if err != nil {
			return fmt.Errorf("error reading dir %q: %w", p, err)
		}
		if len(entries) > 0 {
			return nil
		}
		logboek.Context(ctx).Debug().LogF("Removing empty dir %s\n", p)
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("unable to remove empty dir %q: %w", p, err)
		}
	}
}

func removeInsidePath(ctx context.Context, path string) error {
	stat, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return fmt.Errorf("unable to access path %q: %w", path, err)
	}

	if !stat.IsDir() {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("error reading dir %q: %w", path, err)
	}

	for _, entry := range entries {
		destPath := filepath.Join(path, entry.Name())

		logboek.Context(ctx).Debug().LogF("Removing path %s\n", destPath)
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("unable to remove path %q: %w", destPath, err)
		}
	}

	return nil
}
