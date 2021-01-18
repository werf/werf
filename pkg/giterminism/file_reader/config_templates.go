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
		pattern,
		r.manager.Config().IsUncommittedConfigTemplateFileAccepted,
		r.readCommitConfigTemplateFile,
		func(relPath string, data []byte, err error) error {
			templatePathInsideDir := util.GetRelativeToBaseFilepath(templatesDirRelPath, relPath)
			return tmplFunc(templatePathInsideDir, data, err)
		},
		func(relPaths ...string) error {
			return NewUncommittedFilesError("werf config template", relPaths...)
		},
		func(relPaths ...string) error {
			return NewUncommittedFilesChangesError("werf config template", relPaths...)
		},
	)
}

func (r FileReader) readCommitConfigTemplateFile(ctx context.Context, relPath string) ([]byte, error) {
	return r.readCommitFile(ctx, relPath, func(ctx context.Context, relPath string) error {
		return NewUncommittedFilesChangesError("werf config template", relPath)
	})
}
