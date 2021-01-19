package file_reader

import "context"

const GiterminismConfigName = "werf-giterminism.yaml"

func (r FileReader) IsGiterminismConfigExistAnywhere(ctx context.Context) (bool, error) {
	return r.isConfigurationFileExistAnywhere(ctx, GiterminismConfigName)
}

func (r FileReader) ReadGiterminismConfig(ctx context.Context) ([]byte, error) {
	if err := r.checkConfigurationFileExistence(ctx, giterminismConfigErrorConfigType, GiterminismConfigName, func(relPath string) (bool, error) {
		return false, nil
	}); err != nil {
		return nil, err
	}

	return r.readGiterminismConfig(ctx)
}

func (r FileReader) readGiterminismConfig(ctx context.Context) ([]byte, error) {
	return r.readConfigurationFile(ctx, giterminismConfigErrorConfigType, GiterminismConfigName, func(relPath string) (bool, error) {
		return false, nil
	})
}
