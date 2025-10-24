package filemanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/git_repo"
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

	ListFilesByGlob(ctx context.Context, dir, glob string) ([]string, error)

	IsRegularFileExist(ctx context.Context, relPath string) (exist bool, err error)

	file.ChartFileReaderInterface
}

type FileManager struct {
	fileReader FileReader
	includes   []*includes.Include
	caches     *caches

	customProjectDir string
}

type caches struct {
	dockerFiles map[string][]byte
}

type NewFileManagerOptions struct {
	ProjectDir             string
	FileReader             FileReader
	Inspector              inspector.Inspector
	LocalGitRepo           *git_repo.Local
	CreateIncludesLockFile bool
	AllowIncludesUpdate    bool
}

func NewFileManager(ctx context.Context, opts NewFileManagerOptions) (*FileManager, error) {
	if opts.AllowIncludesUpdate {
		if err := opts.Inspector.InspectIncludesAllowUpdate(); err != nil {
			return nil, err
		}
	}
	includes, err := includes.Init(ctx, includes.InitIncludesOptions{
		FileReader:             opts.FileReader,
		CreateOrUpdateLockFile: opts.CreateIncludesLockFile,
		UseLatestVersion:       opts.AllowIncludesUpdate,
		ProjectDir:             opts.ProjectDir,
		LocalGitRepo:           opts.LocalGitRepo,
	})
	if err != nil {
		return nil, err
	}
	return &FileManager{
		fileReader: opts.FileReader,
		includes:   includes,
		caches: &caches{
			dockerFiles: make(map[string][]byte),
		},
		customProjectDir: opts.ProjectDir,
	}, nil
}

func (f *FileManager) ReadConfig(ctx context.Context, relPath string) (string, []byte, error) {
	exists, _ := f.fileReader.IsConfigExistAnywhere(ctx, relPath)
	if !exists && len(f.includes) > 0 {
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
	werfTemplatesCache := make(map[string]struct{})
	err := f.fileReader.ReadConfigTemplateFiles(ctx, customRelDirPath, func(templatePathInsideDir string, data []byte, err error) error {
		if err != nil {
			return err
		}
		templateName := filepath.ToSlash(templatePathInsideDir)
		werfTemplatesCache[templateName] = struct{}{}

		return tmplFunc(templateName, string(data))
	})
	if err != nil {
		return fmt.Errorf("unable to read werf config templates: %w", err)
	}

	for _, include := range f.includes {
		err := include.WalkObjects(func(toPath, _ string) error {
			normToPath := filepath.ToSlash(toPath)

			if strings.HasPrefix(normToPath, filepath.ToSlash(file_reader.ConfigTemplatesPathList(customRelDirPath))) {

				if _, ok := werfTemplatesCache[normToPath]; ok {
					return nil
				}

				data, err := include.GetFile(ctx, normToPath)
				if err != nil {
					return fmt.Errorf("unable to read included template %q from %s: %w", normToPath, include.GetName(), err)
				}

				if err := tmplFunc(normToPath, string(data)); err != nil {
					return fmt.Errorf("unable to process included template %q: %w", normToPath, err)
				}

				werfTemplatesCache[normToPath] = struct{}{}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileManager) tryReadFromIncludes(ctx context.Context, relPath string) ([]byte, error) {
	for _, include := range f.includes {
		data, err := include.GetFile(ctx, relPath)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("unable to read file %q from includes: not found", relPath)
}

func (f *FileManager) ConfigGoTemplateFilesGet(ctx context.Context, relPath string) ([]byte, error) {
	exists, err := f.fileReader.IsRegularFileExist(ctx, relPath)
	if err != nil {
		return nil, err
	}

	if !exists && len(f.includes) > 0 {
		data, err := f.tryReadFromIncludes(ctx, relPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read file %q: %w", relPath, err)
		}
		return data, nil
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
	if data, ok := f.caches.dockerFiles[relPath]; ok {
		return data, nil
	}

	localPath := getDirAbsPath(relPath, f.customProjectDir)
	if exist, _ := util.FileExists(localPath); exist {
		dockerfileData, err := f.fileReader.ReadDockerfile(ctx, relPath)
		if err != nil {
			return nil, err
		}
		f.caches.dockerFiles[relPath] = dockerfileData
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
			f.caches.dockerFiles[relPath] = dockerfileData
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
		_, err := f.tryReadFromIncludes(ctx, relPath)
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

	data, err := f.tryReadFromIncludes(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read dockerignore file %q: %w", filepath.ToSlash(relPath), err)
	}

	return data, nil
}

func (f *FileManager) LocateChart(ctx context.Context, name string) (string, error) {
	path := getDirAbsPath(name, f.customProjectDir)
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
	relToProjFilePath := util.GetRelativeToBaseFilepath(f.customProjectDir, filePath)

	path := getDirAbsPath(relToProjFilePath, f.customProjectDir)
	if exist, _ := util.FileExists(path); exist {
		data, err := f.fileReader.ReadChartFile(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("unable to read chart file %q: %w", path, err)
		}
		return data, nil
	}

	logboek.Context(ctx).Debug().LogF("Chart file %q not found in the local filesystem. Try find in includes\n", relToProjFilePath)
	data, err := f.tryReadFromIncludes(ctx, relToProjFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read chart file %q: %w", filepath.ToSlash(relToProjFilePath), err)
	}

	return data, nil
}

func (f *FileManager) LoadChartDir(ctx context.Context, dir string) ([]*file.ChartExtenderBufferedFile, error) {
	chartLocalAbsPath := getDirAbsPath(dir, f.customProjectDir)
	processed := make(map[string]bool)

	var chartDir []*file.ChartExtenderBufferedFile

	readFromLocalFs, err := loadChartDirFromLocalSource(chartLocalAbsPath)
	if err != nil {
		return nil, err
	}

	if readFromLocalFs {
		var err error
		chartDir, err = f.fileReader.LoadChartDir(ctx, chartLocalAbsPath)
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
				return fmt.Errorf("unable to read included chart file %q from %s: %w", normToPath, include.GetName(), err)
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
		return nil, fmt.Errorf("load chart dir error: the directory %q not found in the project git repository or includes", dir)
	}

	return chartDir, nil
}

func loadChartDirFromLocalSource(dir string) (bool, error) {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("os.Stat failed for: %w", err)
	}
	return true, nil
}

func getDirAbsPath(relPath, customProjectDir string) string {
	if path.IsAbs(relPath) {
		return relPath
	}
	if customProjectDir != "" {
		relPath = filepath.Join(customProjectDir, relPath)
	}
	return util.GetAbsoluteFilepath(relPath)
}

func (f *FileManager) ChartIsDir(relPath string) (bool, error) {
	chartLocalAbsPath := getDirAbsPath(relPath, f.customProjectDir)
	if fi, err := os.Stat(chartLocalAbsPath); err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("os.Stat failed for %q: %w", chartLocalAbsPath, err)
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
			return false, fmt.Errorf("walkObjects failed for include %q: %w", include.GetName(), err)
		}
	}

	if hasFile {
		return false, nil
	}

	return false, fmt.Errorf("check chart is dir error: path %q not found on local filesystem or includes", relPath)
}

const (
	fromFsSource = "local"
)

func (f *FileManager) ListFilesByGlob(ctx context.Context, sources, globs []string) (map[string]string, error) {
	var fromFs []string
	for _, glob := range globs {
		files, err := f.fileReader.ListFilesByGlob(ctx, "", glob)
		if err != nil {
			return nil, err
		}
		fromFs = append(fromFs, files...)
	}

	fsFilesSet := make(map[string]struct{})
	for _, path := range fromFs {
		fsFilesSet[path] = struct{}{}
	}

	fromIncludes := includes.ListFilesByGlobs(ctx, f.includes, globs, sources)

	result := make(map[string]string)

	if includesFromFs(sources) {
		for _, path := range fromFs {
			result[path] = fromFsSource
		}
	}

	for path, inc := range fromIncludes {
		if _, exists := fsFilesSet[path]; !exists {
			result[path] = inc.GetName()
		}
	}

	return result, nil
}

func includesFromFs(sources []string) bool {
	return len(sources) == 0 || slices.Contains(sources, fromFsSource)
}
