package file_reader

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
)

func (r FileReader) IsDockerignoreExistAnywhere(ctx context.Context, relPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsDockerignoreExistAnywhere %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = r.IsConfigurationFileExistAnywhere(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) ReadDockerfile(ctx context.Context, relPath string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadDockerfile %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readDockerfile(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %d\nerr: %q\n", len(data), err)
			}
		})

	if err != nil {
		return nil, fmt.Errorf("unable to read dockerfile %q: %w", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (r FileReader) readDockerfile(ctx context.Context, relPath string) ([]byte, error) {
	return r.ReadAndCheckConfigurationFile(ctx, relPath, r.giterminismConfig.IsUncommittedDockerfileAccepted)
}

func (r FileReader) ReadDockerignore(ctx context.Context, relPath string) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadDockerignore %q", relPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readDockerignore(ctx, relPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %d\nerr: %q\n", len(data), err)
			}
		})

	if err != nil {
		return nil, fmt.Errorf("unable to read dockerignore file %q: %w", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (r FileReader) readDockerignore(ctx context.Context, relPath string) ([]byte, error) {
	return r.ReadAndCheckConfigurationFile(ctx, relPath, r.giterminismConfig.IsUncommittedDockerignoreAccepted)
}
