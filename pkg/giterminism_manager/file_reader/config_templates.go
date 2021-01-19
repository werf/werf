package file_reader

import (
	"context"
	"path/filepath"

	"github.com/werf/werf/pkg/util"
)

var DefaultWerfConfigTemplatesDirName = ".werf"

func (r FileReader) ReadConfigTemplateFiles(ctx context.Context, customDirRelPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error {
	templatesDirRelPath := DefaultWerfConfigTemplatesDirName
	if customDirRelPath != "" {
		templatesDirRelPath = customDirRelPath
	}

	pattern := filepath.Join(templatesDirRelPath, "**", "*.tmpl")
	return r.configurationFilesGlob(
		ctx,
		configTemplateErrorConfigType,
		pattern,
		r.giterminismConfig.IsUncommittedConfigTemplateFileAccepted,
		r.readCommitConfigTemplateFile,
		func(relPath string, data []byte, err error) error {
			templatePathInsideDir := util.GetRelativeToBaseFilepath(templatesDirRelPath, relPath)
			return tmplFunc(templatePathInsideDir, data, err)
		},
	)
}

func (r FileReader) readCommitConfigTemplateFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.readCommitFile(ctx, relPath, func(ctx context.Context, relPath string) error {
		return NewUncommittedFilesChangesError(configTemplateErrorConfigType, relPath)
	})
}
