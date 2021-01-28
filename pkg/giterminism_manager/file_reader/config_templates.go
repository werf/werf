package file_reader

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/util"
)

var DefaultWerfConfigTemplatesDirName = ".werf"

func (r FileReader) ReadConfigTemplateFiles(ctx context.Context, customDirRelPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error {
	err := r.readConfigTemplateFiles(ctx, customDirRelPath, tmplFunc)
	if err != nil {
		return fmt.Errorf("unable to read werf config templates: %s", err)
	}

	return nil
}

func (r FileReader) readConfigTemplateFiles(ctx context.Context, customDirRelPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error {
	templatesDirRelPath := DefaultWerfConfigTemplatesDirName
	if customDirRelPath != "" {
		templatesDirRelPath = customDirRelPath
	}

	return r.WalkConfigurationFilesWithGlob(
		ctx,
		templatesDirRelPath,
		"**/*.tmpl",
		r.giterminismConfig.IsUncommittedConfigTemplateFileAccepted,
		func(relPath string, data []byte, err error) error {
			templatePathInsideDir := util.GetRelativeToBaseFilepath(templatesDirRelPath, relPath)
			return tmplFunc(templatePathInsideDir, data, err)
		},
	)
}
