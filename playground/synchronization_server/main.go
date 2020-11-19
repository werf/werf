package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/lockgate"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/storage"

	"github.com/werf/werf/pkg/storage/synchronization_server"
)

func doMain() error {
	if err := werf.Init("", ""); err != nil {
		return err
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Init(); err != nil {
		return err
	}

	logboek.Context(ctx).SetLevel(logboek.Context(ctx).Debug)

	synchronizationServerDir := "synchronization-server"

	return synchronization_server.RunSynchronizationServer(
		"localhost", "55581",
		func(clientID string) (storage.LockManager, error) {
			if locker, err := lockgate.NewFileLocker(filepath.Join(synchronizationServerDir, "lock-manager", clientID)); err != nil {
				return nil, err
			} else {
				return storage.NewGenericLockManager(locker), nil
			}
		},
		func(clientID string) (storage.StagesStorageCache, error) {
			return storage.NewFileStagesStorageCache(filepath.Join(synchronizationServerDir, "stages-storage-cache", clientID)), nil
		},
	)
}

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
}
