package file_reader

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
)

const GiterminismConfigName = "werf-giterminism.yaml"

func (r FileReader) IsGiterminismConfigExistAnywhere(ctx context.Context) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsGiterminismConfigExistAnywhere").
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			exist, err = r.IsConfigurationFileExistAnywhere(ctx, GiterminismConfigName)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) ReadGiterminismConfig(ctx context.Context) (data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadGiterminismConfig").
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			data, err = r.readGiterminismConfig(ctx)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	if err != nil {
		return nil, fmt.Errorf("unable to read werf giterminism config: %s", err)
	}

	return
}

func (r FileReader) readGiterminismConfig(ctx context.Context) ([]byte, error) {
	return r.ReadAndCheckConfigurationFile(ctx, GiterminismConfigName, func(relPath string) (bool, error) {
		return false, nil
	})
}
