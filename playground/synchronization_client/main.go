package main

import (
	"fmt"
	"os"
	"time"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/werf"

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

	client := synchronization_server.NewLockManagerHttpClient("http://localhost:55581/lock-manager")
	if lockHandle, err := client.LockStage("myproj", "5050505050"); err != nil {
		return err
	} else {
		time.Sleep(10 * time.Second)
		if err := client.Unlock(lockHandle); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
}
