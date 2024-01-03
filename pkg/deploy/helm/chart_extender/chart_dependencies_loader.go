package chart_extender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/provenance"
	"helm.sh/helm/v3/pkg/registry"
	"sigs.k8s.io/yaml"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

func GetChartDependenciesCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "helm_chart_dependencies", "1")
}

func LoadMetadata(files []*chart.ChartExtenderBufferedFile) (*chart.Metadata, error) {
	var metadata *chart.Metadata

	for _, f := range files {
		if f.Name == "Chart.yaml" {
			metadata = new(chart.Metadata)

			if err := yaml.Unmarshal(f.Data, metadata); err != nil {
				return nil, errors.Wrap(err, "cannot load Chart.yaml")
			}

			if metadata.APIVersion == "" {
				metadata.APIVersion = chart.APIVersionV1
			}

			break
		}
	}

	for _, f := range files {
		if f.Name == "requirements.yaml" {
			if err := yaml.Unmarshal(f.Data, metadata); err != nil {
				return nil, errors.Wrap(err, "cannot load requirements.yaml")
			}
			break
		}
	}

	return metadata, nil
}

func CopyChartDependenciesIntoCache(ctx context.Context, chartDir string, ch *chart.Chart) error {
	metadataBytes, err := yaml.Marshal(ch.Metadata)
	if err != nil {
		return fmt.Errorf("unable to marshal chart metadata into yaml: %w", err)
	}

	metadataLockBytes, err := yaml.Marshal(ch.Lock)
	if err != nil {
		return fmt.Errorf("unable to marshal chart metadata lock into yaml: %w", err)
	}

	_, err = prepareDependenciesDir(ctx, metadataBytes, metadataLockBytes, func(tmpDepsDir string) error {
		var archivesDependencies []*chart.File

	FindArchivesDependencies:
		for _, f := range ch.Raw {
			if strings.HasPrefix(f.Name, "charts/") {
				for _, depLock := range ch.Lock.Dependencies {
					if filepath.Base(f.Name) == MakeDependencyArchiveName(depLock.Name, depLock.Version) {
						archivesDependencies = append(archivesDependencies, f)
						continue FindArchivesDependencies
					}
				}
			}
		}

		for _, depArch := range archivesDependencies {
			srcPath := filepath.Join(chartDir, depArch.Name)
			destPath := filepath.Join(tmpDepsDir, depArch.Name)
			if err := copy.Copy(srcPath, destPath); err != nil {
				return fmt.Errorf("unable to copy %q into cache %q: %w", srcPath, tmpDepsDir, err)
			}
		}

		return nil
	}, logboek.Context(ctx).Info())

	return err
}

func prepareDependenciesDir(ctx context.Context, metadataBytes, metadataLockBytes []byte, prepareFunc func(tmpDepsDir string) error, logger types.ManagerInterface) (string, error) {
	depsDir := filepath.Join(GetChartDependenciesCacheDir(), util.Sha256Hash(string(metadataLockBytes)))

	_, err := os.Stat(depsDir)
	switch {
	case os.IsNotExist(err):
		if err := logger.LogProcess("Preparing chart dependencies").DoError(func() error {
			logger.LogF("Using chart dependencies directory: %s\n", depsDir)
			_, lock, err := werf.AcquireHostLock(ctx, depsDir, lockgate.AcquireOptions{})
			if err != nil {
				return fmt.Errorf("error acquiring lock for %q: %w", depsDir, err)
			}
			defer werf.ReleaseHostLock(lock)

			switch _, err := os.Stat(depsDir); {
			case os.IsNotExist(err):
			case err != nil:
				return fmt.Errorf("error accessing %s: %w", depsDir, err)
			default:
				// at the time we have acquired a lock the target directory was created
				return nil
			}

			tmpDepsDir := fmt.Sprintf("%s.tmp.%s", depsDir, uuid.NewString())

			if err := createChartDependenciesDir(tmpDepsDir, metadataBytes, metadataLockBytes); err != nil {
				return err
			}

			if err := prepareFunc(tmpDepsDir); err != nil {
				return err
			}

			if err := os.Rename(tmpDepsDir, depsDir); err != nil {
				return fmt.Errorf("error renaming %q to %q: %w", tmpDepsDir, depsDir, err)
			}
			return nil
		}); err != nil {
			return "", err
		}
	case err != nil:
		return "", fmt.Errorf("error accessing %q: %w", depsDir, err)
	default:
		logger.LogF("Using cached chart dependencies directory: %s\n", depsDir)
	}

	return depsDir, nil
}

func createChartDependenciesDir(destDir string, metadataBytes, metadataLockBytes []byte) error {
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %w", destDir, err)
	}

	files := []*chart.ChartExtenderBufferedFile{
		{Name: "Chart.yaml", Data: metadataBytes},
		{Name: "Chart.lock", Data: metadataLockBytes},
	}

	for _, file := range files {
		if file == nil {
			continue
		}

		path := filepath.Join(destDir, file.Name)
		if err := ioutil.WriteFile(path, file.Data, 0o644); err != nil {
			return fmt.Errorf("error writing %q: %w", path, err)
		}
	}

	return nil
}

func GetPreparedChartDependenciesDir(ctx context.Context, metadataFile, metadataLockFile *chart.ChartExtenderBufferedFile, helmEnvSettings *cli.EnvSettings, registryClient *registry.Client, buildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions) (string, error) {
	return prepareDependenciesDir(ctx, metadataFile.Data, metadataLockFile.Data, func(tmpDepsDir string) error {
		buildChartDependenciesOpts.LoadOptions = &loader.LoadOptions{
			ChartExtender:               NewWerfChartStub(ctx, buildChartDependenciesOpts.IgnoreInvalidAnnotationsAndLabels),
			SubchartExtenderFactoryFunc: nil,
		}
		if err := command_helpers.BuildChartDependenciesInDir(ctx, tmpDepsDir, helmEnvSettings, registryClient, buildChartDependenciesOpts); err != nil {
			return fmt.Errorf("error building chart dependencies: %w", err)
		}
		return nil
	}, logboek.Context(ctx).Default())
}

type ChartDependenciesConfiguration struct {
	ChartMetadata     *chart.Metadata
	ChartMetadataLock *chart.Lock
}

func NewChartDependenciesConfiguration(chartMetadata *chart.Metadata, chartMetadataLock *chart.Lock) *ChartDependenciesConfiguration {
	return &ChartDependenciesConfiguration{ChartMetadata: chartMetadata, ChartMetadataLock: chartMetadataLock}
}

func (conf *ChartDependenciesConfiguration) GetExternalDependenciesFiles(loadedChartFiles []*chart.ChartExtenderBufferedFile) (bool, *chart.ChartExtenderBufferedFile, *chart.ChartExtenderBufferedFile, error) {
	metadataBytes, err := yaml.Marshal(conf.ChartMetadata)
	if err != nil {
		return false, nil, nil, fmt.Errorf("unable to marshal original chart metadata into yaml: %w", err)
	}
	metadata := new(chart.Metadata)
	if err := yaml.Unmarshal(metadataBytes, metadata); err != nil {
		return false, nil, nil, fmt.Errorf("unable to unmarshal original chart metadata yaml: %w", err)
	}

	metadataLockBytes, err := yaml.Marshal(conf.ChartMetadataLock)
	if err != nil {
		return false, nil, nil, fmt.Errorf("unable to marshal original chart metadata lock into yaml: %w", err)
	}
	metadataLock := new(chart.Lock)
	if err := yaml.Unmarshal(metadataLockBytes, metadataLock); err != nil {
		return false, nil, nil, fmt.Errorf("unable to unmarshal original chart metadata lock yaml: %w", err)
	}

	metadata.APIVersion = "v2"

	var externalDependenciesNames []string
	isExternalDependency := func(depName string) bool {
		for _, externalDepName := range externalDependenciesNames {
			if depName == externalDepName {
				return true
			}
		}
		return false
	}

FindExternalDependencies:
	for _, depLock := range metadataLock.Dependencies {
		if depLock.Repository == "" || strings.HasPrefix(depLock.Repository, "file://") {
			continue
		}

		for _, loadedFile := range loadedChartFiles {
			if strings.HasPrefix(loadedFile.Name, "charts/") {
				if filepath.Base(loadedFile.Name) == MakeDependencyArchiveName(depLock.Name, depLock.Version) {
					continue FindExternalDependencies
				}
			}
		}

		externalDependenciesNames = append(externalDependenciesNames, depLock.Name)
	}

	var filteredLockDependencies []*chart.Dependency
	for _, depLock := range metadataLock.Dependencies {
		if isExternalDependency(depLock.Name) {
			filteredLockDependencies = append(filteredLockDependencies, depLock)
		}
	}
	metadataLock.Dependencies = filteredLockDependencies

	var filteredDependencies []*chart.Dependency
	for _, dep := range metadata.Dependencies {
		if isExternalDependency(dep.Name) {
			filteredDependencies = append(filteredDependencies, dep)
		}
	}
	metadata.Dependencies = filteredDependencies

	if len(metadata.Dependencies) == 0 {
		return false, nil, nil, nil
	}

	// Set resolved repository from the lock file
	for _, dep := range metadata.Dependencies {
		for _, depLock := range metadataLock.Dependencies {
			if dep.Name == depLock.Name {
				dep.Repository = depLock.Repository
				break
			}
		}
	}

	if newDigest, err := HashReq(metadata.Dependencies, metadataLock.Dependencies); err != nil {
		return false, nil, nil, fmt.Errorf("unable to calculate external dependencies Chart.yaml digest: %w", err)
	} else {
		metadataLock.Digest = newDigest
	}

	metadataFile := &chart.ChartExtenderBufferedFile{Name: "Chart.yaml"}
	if data, err := yaml.Marshal(metadata); err != nil {
		return false, nil, nil, fmt.Errorf("unable to marshal chart metadata file with external dependencies: %w", err)
	} else {
		metadataFile.Data = data
	}

	metadataLockFile := &chart.ChartExtenderBufferedFile{Name: "Chart.lock"}
	if data, err := yaml.Marshal(metadataLock); err != nil {
		return false, nil, nil, fmt.Errorf("unable to marshal chart metadata lock file with external dependencies: %w", err)
	} else {
		metadataLockFile.Data = data
	}

	return true, metadataFile, metadataLockFile, nil
}

func LoadChartDependencies(ctx context.Context, loadChartDirFunc func(ctx context.Context, dir string) ([]*chart.ChartExtenderBufferedFile, error), chartDir string, loadedChartFiles []*chart.ChartExtenderBufferedFile, helmEnvSettings *cli.EnvSettings, registryClient *registry.Client, buildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions) ([]*chart.ChartExtenderBufferedFile, error) {
	res := loadedChartFiles

	var chartMetadata *chart.Metadata
	var chartMetadataLock *chart.Lock

	for _, f := range loadedChartFiles {
		switch f.Name {
		case "Chart.yaml":
			chartMetadata = new(chart.Metadata)
			if err := yaml.Unmarshal(f.Data, chartMetadata); err != nil {
				return nil, errors.Wrap(err, "cannot load Chart.yaml")
			}
			if chartMetadata.APIVersion == "" {
				chartMetadata.APIVersion = chart.APIVersionV1
			}

		case "Chart.lock":
			chartMetadataLock = new(chart.Lock)
			if err := yaml.Unmarshal(f.Data, chartMetadataLock); err != nil {
				return nil, errors.Wrap(err, "cannot load Chart.lock")
			}
		}
	}

	for _, f := range loadedChartFiles {
		switch f.Name {
		case "requirements.yaml":
			if chartMetadata == nil {
				chartMetadata = new(chart.Metadata)
			}
			if err := yaml.Unmarshal(f.Data, chartMetadata); err != nil {
				return nil, errors.Wrap(err, "cannot load requirements.yaml")
			}

		case "requirements.lock":
			if chartMetadataLock == nil {
				chartMetadataLock = new(chart.Lock)
			}
			if err := yaml.Unmarshal(f.Data, chartMetadataLock); err != nil {
				return nil, errors.Wrap(err, "cannot load requirements.lock")
			}
		}
	}

	if chartMetadata == nil {
		return res, nil
	}

	if chartMetadataLock == nil {
		if len(chartMetadata.Dependencies) > 0 {
			logboek.Context(ctx).Error().LogLn("Cannot build chart dependencies and preload charts without lock file (.helm/Chart.lock or .helm/requirements.lock)")
			logboek.Context(ctx).Error().LogLn("It is recommended to add Chart.lock file to your project repository or remove chart dependencies.")
			logboek.Context(ctx).Error().LogLn()
			logboek.Context(ctx).Error().LogLn("To generate a lock file run 'werf helm dependency update .helm' and commit resulting .helm/Chart.lock or .helm/requirements.lock (it is not required to commit whole .helm/charts directory, better add it to the .gitignore).")
			logboek.Context(ctx).Error().LogLn()
		}

		return res, nil
	}

	conf := NewChartDependenciesConfiguration(chartMetadata, chartMetadataLock)

	// Append virtually loaded files from custom dependency repositories in the local filesystem,
	// pretending these files are located in the charts/ dir as designed in the Helm.
	for _, chartDep := range chartMetadataLock.Dependencies {
		if !strings.HasPrefix(chartDep.Repository, "file://") {
			continue
		}

		relativeLocalChartPath := strings.TrimPrefix(chartDep.Repository, "file://")
		localChartPath := filepath.Join(chartDir, relativeLocalChartPath)

		chartFiles, err := loadChartDirFunc(ctx, localChartPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load custom subchart dir %q: %w", localChartPath, err)
		}

		for _, f := range chartFiles {
			f.Name = filepath.Join("charts", chartDep.Name, f.Name)
		}

		res = append(res, chartFiles...)
	}

	haveExternalDependencies, metadataFile, metadataLockFile, err := conf.GetExternalDependenciesFiles(loadedChartFiles)
	if err != nil {
		return nil, fmt.Errorf("unable to get external dependencies chart configuration files: %w", err)
	}
	if !haveExternalDependencies {
		return res, nil
	}

	depsDir, err := GetPreparedChartDependenciesDir(ctx, metadataFile, metadataLockFile, helmEnvSettings, registryClient, buildChartDependenciesOpts)
	if err != nil {
		return nil, fmt.Errorf("error preparing chart dependencies: %w", err)
	}
	localFiles, err := loader.GetFilesFromLocalFilesystem(depsDir)
	if err != nil {
		return nil, err
	}

	for _, f := range localFiles {
		if strings.HasPrefix(f.Name, "charts/") {
			f1 := new(chart.ChartExtenderBufferedFile)
			*f1 = chart.ChartExtenderBufferedFile(*f)
			res = append(res, f1)
			logboek.Context(ctx).Debug().LogF("-- LoadChartDependencies: loading subchart %q from the dependencies dir %q\n", f.Name, depsDir)
		}
	}

	return res, nil
}

func HashReq(req, lock []*chart.Dependency) (string, error) {
	data, err := json.Marshal([2][]*chart.Dependency{req, lock})
	if err != nil {
		return "", err
	}
	s, err := provenance.Digest(bytes.NewBuffer(data))
	return "sha256:" + s, err
}

func MakeDependencyArchiveName(depName, depVersion string) string {
	return fmt.Sprintf("%s-%s.tgz", depName, depVersion)
}
