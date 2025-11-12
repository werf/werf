package tmp_manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/samber/lo"
)

var registrator = newGCRegistrator()

type gcRegistrator struct {
	pathQueue []lo.Tuple2[string, string]
	mutex     sync.Mutex
}

func newGCRegistrator() *gcRegistrator {
	return &gcRegistrator{}
}

func (r *gcRegistrator) registerAll(_ context.Context) error {
	for _, item := range r.pathQueue {
		if err := registerPath(item.A, item.B); err != nil {
			return err
		}
	}

	return nil
}

func (r *gcRegistrator) queueRegistration(_ context.Context, actualPath, targetDir string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.pathQueue = append(r.pathQueue, lo.Tuple2[string, string]{
		A: actualPath,
		B: targetDir,
	})

	return nil
}

// DelegateCleanup delegates the cleanup to "werf host cleanup"
func DelegateCleanup(ctx context.Context) error {
	return registrator.registerAll(ctx)
}

func registerPath(actualPath, targetDir string) error {
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %w", targetDir, err)
	}

	targetPath := filepath.Join(targetDir, filepath.Base(actualPath))

	if err := os.Symlink(actualPath, targetPath); err != nil {
		return fmt.Errorf("unable to create symlink %s -> %s: %w", targetPath, actualPath, err)
	}

	return nil
}
