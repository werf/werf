package file_reader

import (
	"context"
	"fmt"
	"strings"

	"github.com/werf/logboek"
)

func (r FileReader) IsIncludesConfigExistAnywhere(ctx context.Context, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsIncludesConfigExistAnywhere").
		Options(applyDebugToLogboek).
		Do(func() {
			exist, err = r.IsConfigurationFileExistAnywhere(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) ReadIncludesConfig(ctx context.Context, relPath string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadIncludesConfig").
		Options(applyDebugToLogboek).
		Do(func() {
			data, err = r.readIncludesConfig(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	if err != nil {
		return nil, fmt.Errorf("unable to read werf giterminism config: %w", err)
	}

	return
}

func (r FileReader) readIncludesConfig(ctx context.Context, relPath string) ([]byte, error) {
	return r.ReadAndCheckConfigurationFile(ctx, relPath, func(_ string) bool {
		return false
	}, func(path string) (bool, error) {
		return r.IsRegularFileExist(ctx, path)
	})
}

func (r FileReader) ReadIncludesLockFile(ctx context.Context, relPath string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadIncludesLockFile").
		Options(applyDebugToLogboek).
		Do(func() {
			data, err = r.readIncludesLockFile(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// TODO: errors.As() not working, somewhere error wrapped incorrectly
			err = fmt.Errorf("%w\n\nConsider generate lock file using `werf includes update` command", err)
		}
		return nil, fmt.Errorf("unable to read werf includes lock file: %w", err)
	}
	return
}

func (r FileReader) readIncludesLockFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.ReadAndCheckConfigurationFile(ctx, relPath, func(_ string) bool {
		return false
	}, func(path string) (bool, error) {
		return r.IsRegularFileExist(ctx, path)
	})
}
