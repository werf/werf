package filemanager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/giterminism_manager/file_reader"
	"github.com/werf/werf/v2/pkg/includes"
)

type FileReader interface {
	IsConfigExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadConfig(ctx context.Context, customRelPath string) (string, []byte, error)
	ReadConfigTemplateFiles(ctx context.Context, customRelDirPath string, tmplFunc func(templatePathInsideDir string, data []byte, err error) error) error
	ConfigGoTemplateFilesExists(ctx context.Context, relPath string) (bool, error)
	ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error)
	ConfigGoTemplateFilesGlob(ctx context.Context, pattern string) (map[string]interface{}, error)
	ConfigGoTemplateFilesIsDir(ctx context.Context, relPath string) (bool, error)
	ReadDockerfile(ctx context.Context, relPath string) ([]byte, error)
	IsDockerignoreExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadDockerignore(ctx context.Context, relPath string) ([]byte, error)

	IsIncludesConfigExistAnywhere(ctx context.Context, relPath string) (bool, error)
	ReadIncludesConfig(ctx context.Context, relPath string) ([]byte, error)

	file.ChartFileReader
}

type FileManager struct {
	fileReader         FileReader
	includes           []*includes.Include
	werfTemplatesCache map[string]bool
}

func NewFileManager(ctx context.Context, fr FileReader) (*FileManager, error) {
	inlcudes, err := includes.Init(ctx, fr, "")
	if err != nil {
		return nil, err
	}
	return &FileManager{
		fileReader:         fr,
		includes:           inlcudes,
		werfTemplatesCache: make(map[string]bool),
	}, nil
}

func (f *FileManager) ReadConfig(ctx context.Context, relPath string) (string, []byte, error) {
	exists, _ := f.fileReader.IsConfigExistAnywhere(ctx, relPath)
	if !exists {
		configPath, configData, includeErr := includes.FindWerfConfig(ctx, f.includes, file_reader.ConfigPathList(relPath))
		if includeErr != nil {
			return "", nil, fmt.Errorf("unable to find config file %q in includes: %w", relPath, includeErr)
		}
		return configPath, configData, nil
	}
	return f.fileReader.ReadConfig(ctx, relPath)
}

func (f *FileManager) ReadConfigTemplateFiles(ctx context.Context, customRelDirPath string, tmplFunc func(templateName string, content string) error) error {
	err := f.fileReader.ReadConfigTemplateFiles(ctx, customRelDirPath, func(templatePathInsideDir string, data []byte, err error) error {
		if err != nil {
			return err
		}
		templateName := filepath.ToSlash(templatePathInsideDir)
		f.werfTemplatesCache[templateName] = true

		return tmplFunc(templateName, string(data))
	})
	if err != nil {
		return fmt.Errorf("unable to read werf config templates: %w", err)
	}

	for _, include := range f.includes {
		err := include.WalkObjects(func(toPath, _ string) error {
			normToPath := filepath.ToSlash(toPath)

			if strings.HasPrefix(normToPath, filepath.ToSlash(file_reader.ConfigTemplatesPathList(customRelDirPath))) {

				if _, ok := f.werfTemplatesCache[normToPath]; ok {
					return nil
				}

				data, err := include.GetFile(ctx, normToPath)
				if err != nil {
					return fmt.Errorf("unable to read included template %q from %s: %w", normToPath, include.Name, err)
				}

				if err := tmplFunc(normToPath, string(data)); err != nil {
					return fmt.Errorf("unable to process included template %q: %w", normToPath, err)
				}

				f.werfTemplatesCache[normToPath] = true
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	fmt.Println("templatePathCache", f.werfTemplatesCache)

	return nil
}

func (f *FileManager) shouldReadWerfTemplateFromIncludes(relPath string) bool {
	return f.werfTemplatesCache[relPath]
}
func (f *FileManager) tryReadFromInludes(ctx context.Context, relPath string) ([]byte, error) {
	for _, include := range f.includes {
		data, err := include.GetFile(ctx, relPath)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("unable to read file %q from includes: not found", relPath)
}

func (f *FileManager) ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error) {
	if f.shouldReadWerfTemplateFromIncludes(relPath) {
		data, inclErr := f.tryReadFromInludes(ctx, relPath)
		if inclErr == nil {
			return data, nil
		}
	}
	res, err := f.fileReader.ConfigGoTemplateFilesGet(ctx, relPath)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (f *FileManager) ConfigGoTemplateFilesGlob(ctx context.Context, pattern string) (map[string]interface{}, error) {
	res, err := f.fileReader.ConfigGoTemplateFilesGlob(ctx, pattern)
	if err != nil {
		return nil, err
	}

	for _, include := range f.includes {
		includeGlobs, err := include.GetFilesByGlob(ctx, pattern)
		if err != nil {
			return nil, fmt.Errorf("unable to get files by glob from includes: %w", err)
		}
		for relPath, data := range includeGlobs {
			if _, ok := res[relPath]; !ok {
				res[relPath] = string(data)
			}
		}
	}
	return res, nil
}

func (f *FileManager) ConfigGoTemplateFilesExists(ctx context.Context, relPath string) (bool, error) {
	exist, err := f.fileReader.ConfigGoTemplateFilesExists(ctx, relPath)
	if err != nil {
		if includes.IsFileExists(ctx, f.includes, relPath) {
			return true, nil
		}
		return false, err
	}
	return exist, nil
}

func (f *FileManager) ConfigGoTemplateFilesIsDir(ctx context.Context, relPath string) (bool, error) {
	exist, err := f.fileReader.ConfigGoTemplateFilesIsDir(ctx, relPath)
	if err != nil {
		if includes.IsDirExists(ctx, f.includes, relPath) {
			return true, nil
		}
		return false, err
	}
	return exist, nil
}

func (f *FileManager) ReadDockerfile(ctx context.Context, relPath string) ([]byte, error) {
	if exist, _ := util.FileExists(util.GetAbsoluteFilepath(relPath)); exist {
		dockerfileData, err := f.fileReader.ReadDockerfile(ctx, relPath)
		if err != nil {
			return nil, err
		}
		return dockerfileData, nil
	}

	logboek.Context(ctx).Debug().LogF("Dockerfile %q not found in the local filesystem. Try find in includes\n", relPath)

	for _, i := range f.includes {
		if i == nil {
			continue
		}
		dockerfileData, err := i.GetFile(ctx, relPath)
		if err == nil {
			logboek.Context(ctx).Debug().LogF("Found dockerfile %q in includes\n", relPath)
			return dockerfileData, nil
		}
	}
	return nil, fmt.Errorf("unable to read dockerfile %q: file not found", filepath.ToSlash(relPath))
}

func (f *FileManager) IsDockerignoreExistAnywhere(ctx context.Context, relPath string) (bool, error) {
	logboek.Context(ctx).Debug().LogLn("Dockerignore will not be read from includes")
	return f.fileReader.IsDockerignoreExistAnywhere(ctx, relPath)
}
func (f *FileManager) ReadDockerignore(ctx context.Context, relPath string) ([]byte, error) {
	logboek.Context(ctx).Debug().LogLn("Dockerignore will not be read from includes")
	return f.fileReader.ReadDockerignore(ctx, relPath)
}
