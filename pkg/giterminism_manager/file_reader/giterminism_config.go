package file_reader

import (
	"context"
	"fmt"
)

const GiterminismConfigName = "werf-giterminism.yaml"

func (r FileReader) IsGiterminismConfigExistAnywhere(ctx context.Context) (exist bool, err error) {
	return r.IsConfigurationFileExistAnywhere(ctx, GiterminismConfigName)
}

func (r FileReader) ReadGiterminismConfig(ctx context.Context) ([]byte, error) {
	data, err := r.readGiterminismConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to read werf giterminism config: %s", err)
	}

	return data, nil
}

func (r FileReader) readGiterminismConfig(ctx context.Context) ([]byte, error) {
	if err := r.CheckConfigurationFileExistence(ctx, GiterminismConfigName, func(relPath string) (bool, error) {
		return false, nil
	}); err != nil {
		return nil, err
	}

	return r.ReadAndValidateConfigurationFile(ctx, GiterminismConfigName, func(relPath string) (bool, error) {
		return false, nil
	})
}
