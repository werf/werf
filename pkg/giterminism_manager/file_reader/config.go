package file_reader

import (
	"context"
)

var DefaultWerfConfigNames = []string{"werf.yaml", "werf.yml"}

func (r FileReader) IsConfigExistAnywhere(ctx context.Context, customRelPath string) (bool, error) {
	configRelPathList := r.configPathList(customRelPath)
	for _, configRelPath := range configRelPathList {
		if exist, err := r.isConfigurationFileExistAnywhere(ctx, configRelPath); err != nil {
			return false, err
		} else if exist {
			return true, nil
		}
	}

	return false, nil
}

func (r FileReader) ReadConfig(ctx context.Context, customRelPath string) ([]byte, error) {
	configRelPathList := r.configPathList(customRelPath)

	for _, configPath := range configRelPathList {
		if exist, err := r.isConfigExist(ctx, configPath); err != nil {
			return nil, err
		} else if !exist {
			continue
		}

		return r.readConfig(ctx, configPath)
	}

	return nil, r.prepareConfigNotFoundError(ctx, configRelPathList)
}

func (r FileReader) isConfigExist(ctx context.Context, relPath string) (bool, error) {
	return r.isConfigurationFileExist(ctx, relPath, func(_ string) (bool, error) {
		return r.giterminismConfig.IsUncommittedConfigAccepted(), nil
	})
}

func (r FileReader) readConfig(ctx context.Context, relPath string) ([]byte, error) {
	return r.readConfigurationFile(ctx, configErrorConfigType, relPath, func(relPath string) (bool, error) {
		return r.giterminismConfig.IsUncommittedConfigAccepted(), nil
	})
}

func (r FileReader) readCommitConfig(ctx context.Context, relPath string) ([]byte, error) {
	return r.readCommitFile(ctx, relPath, func(ctx context.Context, relPath string) error {
		return NewUncommittedFilesChangesError(configErrorConfigType, relPath)
	})
}

func (r FileReader) prepareConfigNotFoundError(ctx context.Context, configPathsToCheck []string) error {
	for _, configPath := range configPathsToCheck {
		err := r.checkConfigurationFileExistence(ctx, configErrorConfigType, configPath, func(_ string) (bool, error) {
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
		return NewFilesNotFoundInTheProjectDirectoryError(configErrorConfigType, configPath)
	} else {
		return NewFilesNotFoundInTheProjectGitRepositoryError(configErrorConfigType, configPath)
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
