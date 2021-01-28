package file_reader

import (
	"context"
	"fmt"
)

var DefaultWerfConfigNames = []string{"werf.yaml", "werf.yml"}

func (r FileReader) IsConfigExistAnywhere(ctx context.Context, customRelPath string) (bool, error) {
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

func (r FileReader) ReadConfig(ctx context.Context, customRelPath string) ([]byte, error) {
	data, err := r.readConfig(ctx, customRelPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read werf config: %s", err)
	}

	return data, nil
}

func (r FileReader) readConfig(ctx context.Context, customRelPath string) ([]byte, error) {
	configRelPathList := r.configPathList(customRelPath)

	for _, configPath := range configRelPathList {
		if exist, err := r.IsConfigurationFileExist(ctx, configPath, func(_ string) (bool, error) {
			return r.giterminismConfig.IsUncommittedConfigAccepted(), nil
		}); err != nil {
			return nil, err
		} else if !exist {
			continue
		}

		return r.ReadAndValidateConfigurationFile(ctx, configPath, func(_ string) (bool, error) {
			return r.giterminismConfig.IsUncommittedConfigAccepted(), nil
		})
	}

	return nil, r.prepareConfigNotFoundError(ctx, configRelPathList)
}

func (r FileReader) prepareConfigNotFoundError(ctx context.Context, configPathsToCheck []string) (err error) {
	for _, configPath := range configPathsToCheck {
		err := r.CheckConfigurationFileExistence(ctx, configPath, func(_ string) (bool, error) {
			return r.giterminismConfig.IsUncommittedConfigAccepted(), nil
		})

		switch err.(type) {
		case UncommittedFilesError, UncommittedFilesChangesError:
			return err
		}
	}

	var configPath string
	if len(configPathsToCheck) == 1 {
		configPath = configPathsToCheck[0]
	} else { // default werf config (werf.yaml, werf.yml)
		configPath = "werf.yaml"
	}

	if r.sharedOptions.LooseGiterminism() {
		return NewFilesNotFoundInProjectDirectoryError(configPath)
	} else {
		return NewFilesNotFoundInProjectGitRepositoryError(configPath)
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
