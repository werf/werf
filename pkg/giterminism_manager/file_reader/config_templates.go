package file_reader

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
)

var DefaultWerfConfigTemplatesDirName = ".werf"

func (r FileReader) ReadConfigTemplateFiles(ctx context.Context, customDirRelPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) (err error) {
	logboek.Context(ctx).Debug().
		LogBlock("ReadConfigTemplateFiles %q", customDirRelPath).
		Options(func(options types.LogBlockOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			err = r.readConfigTemplateFiles(ctx, customDirRelPath, tmplFunc)

			if debug() {
				logboek.Context(ctx).Debug().LogF("err: %q\n", err)
			}
		})

	if err != nil {
		return fmt.Errorf("unable to read werf config templates: %w", err)
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
		r.giterminismConfig.UncommittedConfigTemplateFilePathMatcher(),
		func(relativeToDirNotResolvedPath string, data []byte, err error) error {
			return tmplFunc(filepath.ToSlash(relativeToDirNotResolvedPath), data, err)
		},
	)
}
