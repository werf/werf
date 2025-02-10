package file_reader

import (
	"context"
	"errors"
	"fmt"

	"github.com/werf/logboek"
)

var DefaultWerfConfigNames = []string{"werf.yaml", "werf.yml"}

func (r FileReader) IsConfigExistAnywhere(ctx context.Context, customRelPath string) (exist bool, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("IsConfigExistAnywhere %q", customRelPath).
		Options(applyDebugToLogboek).
		Do(func() {
			exist, err = r.isConfigExistAnywhere(ctx, customRelPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("exist: %v\nerr: %q\n", exist, err)
			}
		})

	return
}

func (r FileReader) isConfigExistAnywhere(ctx context.Context, customRelPath string) (bool, error) {
	configRelPathList := r.configPathList(customRelPath)
	for _, configRelPath := range configRelPathList {
		if exist, err := r.IsConfigurationFileExistAnywhere(ctx, configRelPath); err != nil {
			return false, err
		} else if exist {
			return true, nil
		}
	}

	return false, nil
}

func (r FileReader) ReadConfig(ctx context.Context, customRelPath string) (path string, data []byte, err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadConfig %q", customRelPath).
		Options(applyDebugToLogboek).
		Do(func() {
			path, data, err = r.readConfig(ctx, customRelPath)

			if debug() {
				logboek.Context(ctx).Debug().LogF("dataLength: %v\nerr: %q\n", len(data), err)
			}
		})

	if err != nil {
		return "", nil, fmt.Errorf("unable to read werf config: %w", err)
	}

	return path, data, nil
}

func (r FileReader) readConfig(ctx context.Context, customRelPath string) (string, []byte, error) {
	configRelPathList := r.configPathList(customRelPath)

	for _, configPath := range configRelPathList {
		data, err := r.ReadAndCheckConfigurationFile(ctx, configPath, func(_ string) bool {
			return r.giterminismConfig.IsUncommittedConfigAccepted()
		}, func(path string) (bool, error) {
			return r.IsRegularFileExist(ctx, path)
		})
		if err != nil {
			if errors.As(err, &FileNotFoundInProjectDirectoryError{}) ||
				errors.As(err, &FileNotFoundInProjectRepositoryError{}) {
				continue
			}

			return "", nil, err
		}

		return configPath, data, nil
	}

	return "", nil, r.PrepareConfigNotFoundError(ctx, configRelPathList)
}

func (r FileReader) PrepareConfigNotFoundError(ctx context.Context, configPathsToCheck []string) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("PrepareConfigNotFoundError %v", configPathsToCheck).
		Options(applyDebugToLogboek).
		Do(func() {
			err = r.prepareConfigNotFoundError(configPathsToCheck)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	return
}

func (r FileReader) prepareConfigNotFoundError(configPathsToCheck []string) error {
	var configPath string
	if len(configPathsToCheck) == 1 {
		configPath = configPathsToCheck[0]
	} else { // default werf config (werf.yaml, werf.yml)
		configPath = "werf.yaml"
	}

	if r.sharedOptions.LooseGiterminism() {
		return r.NewFileNotFoundInProjectDirectoryError(configPath)
	} else {
		return r.NewFileNotFoundInProjectRepositoryError(configPath)
	}
}

func (r FileReader) configPathList(customRelPath string) []string {
	var configRelPathList []string
	if customRelPath != "" {
		configRelPathList = append(configRelPathList, customRelPath)
	} else {
		configRelPathList = DefaultWerfConfigNames
	}

	return configRelPathList
}
