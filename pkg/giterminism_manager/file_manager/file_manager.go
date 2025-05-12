package filemanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/giterminism_manager/file_reader"
	"github.com/werf/werf/v2/pkg/giterminism_manager/inspector"
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
	ReadIncludesLockFile(ctx context.Context, relPath string) (data []byte, err error)

	file.ChartFileReader
}

type FileManager struct {
	fileReader         FileReader
	includes           []*includes.Include
	werfTemplatesCache map[string]bool
}

type NewFileManagerOptions struct {
	FileReader               FileReader
	Inspector                inspector.Inspector
	CreateIncludesLockFile   bool
	UseInludesLatestVersions bool
}

func NewFileManager(ctx context.Context, opts NewFileManagerOptions) (*FileManager, error) {
	if err := opts.Inspector.InspectIncludes(opts.UseInludesLatestVersions); err != nil {
		return nil, fmt.Errorf("includes inspection failed: %w", err)
	}
	inlcudes, err := includes.Init(ctx, includes.InitIncludesOptions{
		FileReader:             opts.FileReader,
		CreateOrUpdateLockFile: opts.CreateIncludesLockFile,
		UseLatestVersion:       opts.UseInludesLatestVersions,
	})
	if err != nil {
		return nil, err
	}
	return &FileManager{
		fileReader:         opts.FileReader,
		includes:           inlcudes,
		werfTemplatesCache: make(map[string]bool),
	}, nil
}

func (f *FileManager) ReadConfig(ctx context.Context, relPath string) (string, []byte, error) {
	exists, _ := f.fileReader.IsConfigExistAnywhere(ctx, relPath)
	if !exists {
		configPath, configData, includeErr := includes.FindWerfConfig(ctx, f.includes, file_reader.ConfigPathList(relPath))
		if includeErr != nil {
			if errors.Is(includeErr, includes.ErrConfigFileNotFound) {
				return "", nil, fmt.Errorf("unable to find any config files %v: %w", file_reader.ConfigPathList(relPath), includeErr)
			}
			return "", nil, fmt.Errorf("unable to read config file %q from includes: %w", relPath, includeErr)
		}
		return configPath, configData, nil
	}
	return f.fileReader.ReadConfig(ctx, relPath)
}

func (f *FileManager) ReadConfigTemplateFiles(ctx context.Context, customRelDirPath string, tmplFunc func(templateName, content string) error) error {
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
	exists, err := f.fileReader.IsDockerignoreExistAnywhere(ctx, relPath)
	if err != nil {
		return false, fmt.Errorf("unable to check dockerignore existence: %w", err)
	}
	if !exists {
		_, err := f.tryReadFromInludes(ctx, relPath)
		if err == nil {
			return true, nil
		}
	}
	return exists, nil
}

func (f *FileManager) ReadDockerignore(ctx context.Context, relPath string) ([]byte, error) {
	exists, err := f.fileReader.IsDockerignoreExistAnywhere(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to check dockerignore existence: %w", err)
	}

	if exists {
		data, err := f.fileReader.ReadDockerignore(ctx, relPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read dockerignore file %q: %w", filepath.ToSlash(relPath), err)
		}
		return data, nil
	}

	data, err := f.tryReadFromInludes(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read dockerignore file %q: %w", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (f *FileManager) LocateChart(ctx context.Context, name string) (string, error) {
	path := util.GetAbsoluteFilepath(name)
	chartDir, err := f.fileReader.LocateChart(ctx, path)
	if err != nil {
		logboek.Context(ctx).Debug().LogF("Chart directory %q not found in the local filesystem. Try find in includes\n", name)
		if includes.IsDirExists(ctx, f.includes, name) {
			return name, nil
		}
		return "", fmt.Errorf("unable to locate chart directory: %w", err)
	}
	return chartDir, nil
}

func (f *FileManager) ReadChartFile(ctx context.Context, filePath string) ([]byte, error) {
	path := util.GetAbsoluteFilepath(filePath)
	if exist, _ := util.FileExists(path); exist {
		data, err := f.fileReader.ReadChartFile(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("unable to read chart file %q: %w", path, err)
		}
		return data, nil
	}

	logboek.Context(ctx).Debug().LogF("Chart file %q not found in the local filesystem. Try find in includes\n", filePath)
	data, err := f.tryReadFromInludes(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read chart file %q: %w", filepath.ToSlash(filePath), err)
	}

	return data, nil
}

func (f *FileManager) LoadChartDir(ctx context.Context, dir string) ([]*file.ChartExtenderBufferedFile, error) {
	absDir := util.GetAbsoluteFilepath(dir)
	processed := make(map[string]bool)

	var chartDir []*file.ChartExtenderBufferedFile

	if exist, _ := util.DirExists(absDir); exist {
		var err error
		chartDir, err = f.fileReader.LoadChartDir(ctx, absDir)
		if err != nil {
			return nil, fmt.Errorf("unable to load chart directory: %w", err)
		}
		for _, file := range chartDir {
			processedPath := filepath.ToSlash(filepath.Join(dir, file.Name))
			logboek.Context(ctx).Debug().LogF("--- %s read from filesystem \n", processedPath)
			processed[processedPath] = true
		}
	}

	logboek.Context(ctx).Debug().LogF("Try to read additional files from includes\n")

	for _, include := range f.includes {
		err := include.WalkObjects(func(toPath, _ string) error {
			normToPath := filepath.ToSlash(toPath)
			normDir := filepath.ToSlash(dir)
			if !strings.HasPrefix(normToPath, normDir+"/") && normToPath != normDir {
				return nil
			}

			if _, ok := processed[normToPath]; ok {
				return nil
			}

			data, err := include.GetFile(ctx, normToPath)
			if err != nil {
				return fmt.Errorf("unable to read included chart file %q from %s: %w", normToPath, include.Name, err)
			}

			logboek.Context(ctx).Debug().LogF("--- %s read from includes \n", normToPath)

			chartDir = append(chartDir, &file.ChartExtenderBufferedFile{
				Name: strings.TrimPrefix(normToPath, normDir+"/"),
				Data: data,
			})
			processed[normToPath] = false
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("unable to walk includes for chart dir: %w", err)
		}
	}

	if len(chartDir) == 0 {
		return nil, fmt.Errorf("the directory %q not found in the project git repository or includes", dir)
	}

	return chartDir, nil
}

func (f *FileManager) ChartIsDir(relPath string) (bool, error) {
	absPath := util.GetAbsoluteFilepath(relPath)

	if fi, err := os.Stat(absPath); err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("os.Stat failed for %q: %w", absPath, err)
	}

	normRelPath := filepath.ToSlash(relPath)
	hasFile := false

	for _, include := range f.includes {
		err := include.WalkObjects(func(toPath, _ string) error {
			normToPath := filepath.ToSlash(toPath)

			if normToPath == normRelPath {
				hasFile = true
			} else if strings.HasPrefix(normToPath, normRelPath+"/") {
				// has files - hence directory
				return errors.New("foundDir")
			}
			return nil
		})
		if err != nil {
			if err.Error() == "foundDir" {
				return true, nil
			}
			return false, fmt.Errorf("walkObjects failed for include %q: %w", include.Name, err)
		}
	}

	if hasFile {
		return false, nil
	}

	return false, fmt.Errorf("path %q not found in local filesystem or includes", relPath)
}
