package bundle_file_reader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/common-go/pkg/util"
)

// BundleFileReader - простая реализация ChartFileReaderInterface для bundle команд
// Не требует Git репозиторий, работает только с локальной файловой системой
type BundleFileReader struct {
	bundleDir string
}

// NewBundleFileReader создает новый BundleFileReader для указанной директории bundle
func NewBundleFileReader(bundleDir string) *BundleFileReader {
	return &BundleFileReader{bundleDir: bundleDir}
}

// LocateChart ищет директорию чарта
func (r *BundleFileReader) LocateChart(ctx context.Context, name string) (string, error) {
	// Для bundle команд чарт всегда находится в bundleDir
	chartDir := r.bundleDir
	if !filepath.IsAbs(name) {
		chartDir = filepath.Join(r.bundleDir, name)
	} else {
		chartDir = name
	}

	// Проверяем существование директории
	if exists, _ := util.DirExists(chartDir); !exists {
		return "", fmt.Errorf("chart directory %q not found", chartDir)
	}

	return chartDir, nil
}

// ReadChartFile читает файл чарта
func (r *BundleFileReader) ReadChartFile(ctx context.Context, filePath string) ([]byte, error) {
	var fullPath string

	if filepath.IsAbs(filePath) {
		fullPath = filePath
	} else {
		fullPath = filepath.Join(r.bundleDir, filePath)
	}

	// Проверяем существование файла
	if exists, _ := util.FileExists(fullPath); !exists {
		return nil, fmt.Errorf("chart file %q not found", fullPath)
	}

	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read chart file %q: %w", fullPath, err)
	}

	return data, nil
}

// LoadChartDir загружает все файлы из директории чарта
func (r *BundleFileReader) LoadChartDir(ctx context.Context, chartDir string) ([]*file.ChartExtenderBufferedFile, error) {
	var fullDir string
	if filepath.IsAbs(chartDir) {
		fullDir = chartDir
	} else {
		fullDir = filepath.Join(r.bundleDir, chartDir)
	}

	var files []*file.ChartExtenderBufferedFile

	err := filepath.Walk(fullDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Получаем относительный путь от fullDir
		relPath, err := filepath.Rel(fullDir, path)
		if err != nil {
			return err
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("unable to read file %q: %w", path, err)
		}

		files = append(files, &file.ChartExtenderBufferedFile{
			Name: relPath,
			Data: data,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to load chart directory %q: %w", fullDir, err)
	}

	return files, nil
}

// ChartIsDir проверяет является ли указанный путь директорией
// Этот метод существует только для обратной совместимости с интерфейсом Loader
func (r *BundleFileReader) ChartIsDir(relPath string) (bool, error) {
	var fullPath string
	if filepath.IsAbs(relPath) {
		fullPath = relPath
	} else {
		fullPath = filepath.Join(r.bundleDir, relPath)
	}

	fi, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("unable to stat %q: %w", fullPath, err)
	}

	return fi.IsDir(), nil
}
