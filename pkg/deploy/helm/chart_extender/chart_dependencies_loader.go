package chart_extender

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/util"

	"github.com/werf/werf/pkg/deploy/helm/command_helpers"

	uuid "github.com/satori/go.uuid"
	"github.com/werf/lockgate"

	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/pkg/errors"
	"github.com/werf/werf/pkg/werf"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"

	"helm.sh/helm/v3/pkg/chart/loader"
)

func GetChartDependenciesCacheDir(lockChecksum string) string {
	return filepath.Join(werf.GetLocalCacheDir(), "helm_chart_dependencies", "1", lockChecksum)
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

func GetPreparedChartDependenciesDir(ctx context.Context, conf *ChartDependenciesConfiguration, helmEnvSettings *cli.EnvSettings, registryClientHandle *helm_v3.RegistryClientHandle, buildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions) (string, error) {
	var lockFileData []byte
	if conf.ChartLockFile != nil {
		lockFileData = conf.ChartLockFile.Data
	} else if conf.RequirementsLockFile != nil {
		lockFileData = conf.RequirementsLockFile.Data
	}

	depsDir := GetChartDependenciesCacheDir(util.Sha256Hash(string(lockFileData)))

	if _, err := os.Stat(depsDir); os.IsNotExist(err) {
		if err := logboek.Context(ctx).Default().LogProcess("Building chart dependencies").DoError(func() error {
			logboek.Context(ctx).Default().LogF("Using chart dependencies directory: %s\n", depsDir)
			if _, lock, err := werf.AcquireHostLock(ctx, depsDir, lockgate.AcquireOptions{}); err != nil {
				return fmt.Errorf("error acquiring lock for %q: %s", depsDir, err)
			} else {
				defer werf.ReleaseHostLock(lock)
			}

			tmpDepsDir := fmt.Sprintf("%s.tmp.%s", depsDir, uuid.NewV4().String())

			buildChartDependenciesOpts.LoadOptions = &loader.LoadOptions{
				ChartExtender:               NewWerfChartStub(ctx),
				SubchartExtenderFactoryFunc: nil,
			}

			if err := command_helpers.BuildChartDependenciesInDir(ctx, conf.ChartFile, conf.ChartLockFile, conf.RequirementsFile, conf.RequirementsLockFile, tmpDepsDir, helmEnvSettings, registryClientHandle, buildChartDependenciesOpts); err != nil {
				return fmt.Errorf("error building chart dependencies: %s", err)
			}

			if err := os.Rename(tmpDepsDir, depsDir); err != nil {
				return fmt.Errorf("error renaming %q to %q: %s", tmpDepsDir, depsDir, err)
			}

			return nil
		}); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", fmt.Errorf("error accessing %q: %s", depsDir, err)
	} else {
		logboek.Context(ctx).Default().LogF("Using cached chart dependencies directory: %s\n", depsDir)
	}

	return depsDir, nil
}

type ChartDependenciesConfiguration struct {
	ChartFile            *chart.ChartExtenderBufferedFile
	ChartLockFile        *chart.ChartExtenderBufferedFile
	RequirementsFile     *chart.ChartExtenderBufferedFile
	RequirementsLockFile *chart.ChartExtenderBufferedFile

	Metadata *chart.Metadata
}

func LoadChartDependencies(ctx context.Context, loadedFiles []*chart.ChartExtenderBufferedFile, helmEnvSettings *cli.EnvSettings, registryClientHandle *helm_v3.RegistryClientHandle, buildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions) ([]*chart.ChartExtenderBufferedFile, error) {
	conf := &ChartDependenciesConfiguration{}

	for _, f := range loadedFiles {
		switch f.Name {
		case "Chart.yaml":
			conf.ChartFile = f

			conf.Metadata = new(chart.Metadata)
			if err := yaml.Unmarshal(f.Data, conf.Metadata); err != nil {
				return nil, errors.Wrap(err, "cannot load Chart.yaml")
			}
			if conf.Metadata.APIVersion == "" {
				conf.Metadata.APIVersion = chart.APIVersionV1
			}

		case "Chart.lock":
			conf.ChartLockFile = f
		case "requirements.lock":
			conf.RequirementsLockFile = f
		}
	}

	for _, f := range loadedFiles {
		switch f.Name {
		case "requirements.yaml":
			conf.RequirementsFile = f

			if conf.Metadata == nil {
				conf.Metadata = new(chart.Metadata)
			}
			if err := yaml.Unmarshal(f.Data, conf.Metadata); err != nil {
				return nil, errors.Wrap(err, "cannot load requirements.yaml")
			}
		}
	}

	if conf.ChartFile == nil {
		return loadedFiles, nil
	}

	if conf.ChartLockFile == nil && conf.RequirementsLockFile == nil {
		if len(conf.Metadata.Dependencies) > 0 {
			logboek.Context(ctx).Error().LogLn("Cannot build chart dependencies and preload charts without lock file (.helm/Chart.lock or .helm/requirements.lock)")
			logboek.Context(ctx).Error().LogLn("It is recommended to add Chart.lock file to your project repository or remove chart dependencies.")
			logboek.Context(ctx).Error().LogLn()
			logboek.Context(ctx).Error().LogLn("To generate a lock file run 'werf helm dependency update .helm' and commit resulting .helm/Chart.lock or .helm/requirements.lock (it is not required to commit whole .helm/charts directory, better add it to the .gitignore).")
			logboek.Context(ctx).Error().LogLn()
		}

		return loadedFiles, nil
	}

	depsDir, err := GetPreparedChartDependenciesDir(ctx, conf, helmEnvSettings, registryClientHandle, buildChartDependenciesOpts)
	if err != nil {
		return nil, fmt.Errorf("error preparing chart dependencies: %s", err)
	}
	localFiles, err := loader.GetFilesFromLocalFilesystem(depsDir)
	if err != nil {
		return nil, err
	}

	res := loadedFiles

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
