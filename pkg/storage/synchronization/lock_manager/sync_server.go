package lock_manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gookit/color"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/storage"
)

var ErrNoSyncServerFound = errors.New("no synchronization server found")

// go build -ldflags github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager.ForceSyncServerRepo=true"
var ForceSyncServerRepo string

// Get sync server name from repository or try to create record is doesn't exist
func GetOrCreateSyncServer(ctx context.Context, projectName, serverAddress string, stagesStorage storage.StagesStorage) (string, error) {
	server, err := getSyncServer(ctx, projectName, stagesStorage)
	if err != nil {
		if errors.Is(err, ErrNoSyncServerFound) {
			createErr := CreateSyncServerRecord(ctx, projectName, serverAddress, stagesStorage)
			if createErr != nil {
				return "", fmt.Errorf("can't create synchronization server record: %w", err)
			}
			return serverAddress, nil

		}
		return "", fmt.Errorf("can't get synchronization server record: %w", err)
	}
	return server, nil
}

func getSyncServer(ctx context.Context, projectName string, stagesStorage storage.StagesStorage) (string, error) {
	server, err := getSyncServerFromStorage(ctx, projectName, stagesStorage)
	if err != nil {
		return "", err
	}

	if server != "" {
		logboek.Context(ctx).Debug().LogF("getSyncServerFromStorage %s selected server: %s\n", projectName, server)
		return server, nil
	}

	return "", ErrNoSyncServerFound
}

// Create sync server name from repository or try to create record is doesn't exist
func CreateSyncServerRecord(ctx context.Context, projectName, serverAddress string, stagesStorage storage.StagesStorage) error {
	now := time.Now()
	timestampMillisec := now.Unix()*1000 + now.UnixNano()/1000_000
	rec := &storage.SyncServerRecord{Server: serverAddress, TimestampMillisec: timestampMillisec}

	_, err := getSyncServer(ctx, projectName, stagesStorage)
	if err != nil {
		if errors.Is(err, ErrNoSyncServerFound) {
			logboek.Context(ctx).Debug().LogF("СreateSyncServerRecord no syncserver found. Creating: %s\n", serverAddress)
			if err := stagesStorage.PostSyncServerRecord(ctx, projectName, rec); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	logboek.Context(ctx).Info().LogF(fmt.Sprintf("СreateSyncServerRecord server already exists. %s will be used\n", serverAddress))

	return nil
}

// Overwrite sync existed sync server
func OverwriteSyncServerRepo(ctx context.Context, projectName, serverAddress string, stagesStorage storage.StagesStorage) error {
	now := time.Now()
	timestampMillisec := now.Unix()*1000 + now.UnixNano()/1000_000
	rec := &storage.SyncServerRecord{Server: serverAddress, TimestampMillisec: timestampMillisec}

	logboek.Context(ctx).Debug().LogF("СreateSyncServerRecord no synchronization server found. Creating: %s\n", serverAddress)
	if err := stagesStorage.PostSyncServerRecord(ctx, projectName, rec); err != nil {
		return fmt.Errorf("unable to overwrite sync server: %w", err)
	}

	return nil
}

func getSyncServerFromStorage(ctx context.Context, projectName string, stagesStorage storage.StagesStorage) (string, error) {
	syncServerReords, err := stagesStorage.GetSyncServerRecords(ctx, projectName)
	if err != nil {
		return "", fmt.Errorf("can't get synchronization server records: %w", err)
	}

	if len(syncServerReords) > 0 {
		res := selectOldestSyncServerRecord(syncServerReords)
		logboek.Context(ctx).Debug().LogF("GetSyncServerFromStorage %s selected server: %s\n", projectName, res.Server)
		return res.Server, nil
	}

	return "", nil
}

func selectOldestSyncServerRecord(records []*storage.SyncServerRecord) *storage.SyncServerRecord {
	var foundRec *storage.SyncServerRecord
	for _, rec := range records {
		if foundRec == nil || rec.TimestampMillisec < foundRec.TimestampMillisec {
			foundRec = rec
		}
	}
	return foundRec
}

// It shows prompt message. Could be used if sync server differs from specified
func PromptRewriteSyncRepoServer(ctx context.Context, specified, repoServer string) error {
	timeout := 2 * time.Minute // magic number
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		prompt := fmt.Sprintf("Specified synchronization server %s is different from synchronization server from repo %s. Do you want to proceed? (Y/n): ", specified, repoServer)
		resp, err := askForConfirmation(prompt)
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()

	select {
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("input timeout: no response within %s minutes. aborted", timeout.String())
	case err := <-errCh:
		return fmt.Errorf("error getting prompt response: %w", err)
	case response := <-respCh:
		if !response {
			return fmt.Errorf("operation aborted. please use the synchronization server from the repository or overwrite the server from the repository. please consider the consequences and proceed at your own risk.")
		}
	}

	return nil
}

func askForConfirmation(prompt string) (bool, error) {
	r := os.Stdin

	fmt.Println(logboek.Colorize(color.Style{color.FgLightYellow}, prompt))

	isTerminal := terminal.IsTerminal(int(r.Fd()))
	if isTerminal {
		if oldState, err := terminal.MakeRaw(int(r.Fd())); err != nil {
			return false, err
		} else {
			defer terminal.Restore(int(r.Fd()), oldState)
		}
	}

	var buf [1]byte
	n, err := r.Read(buf[:])
	if n > 0 {
		switch buf[0] {
		case 'y', 'Y', 13:
			return true, nil
		default:
			return false, nil
		}
	}

	if err != nil && err != io.EOF {
		return false, err
	}

	return false, nil
}
