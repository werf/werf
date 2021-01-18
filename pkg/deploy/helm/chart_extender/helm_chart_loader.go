package chart_extender

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/giterminism"

	"github.com/werf/werf/pkg/deploy/helm/command_helpers"

	uuid "github.com/satori/go.uuid"
	"github.com/werf/lockgate"

	"helm.sh/helm/v3/pkg/cli"

	"github.com/pkg/errors"
	"github.com/werf/werf/pkg/werf"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"

	"helm.sh/helm/v3/pkg/chart/loader"
)

func GetChartDependenciesCacheDir(lockDigest string) string {
	return filepath.Join(werf.GetLocalCacheDir(), "helm_chart_dependencies", lockDigest)
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

func LoadLock(files []*chart.ChartExtenderBufferedFile) (*chart.Lock, *chart.ChartExtenderBufferedFile, error) {
	var lock *chart.Lock
	var lockFile *chart.ChartExtenderBufferedFile

	for _, f := range files {
		switch {
		case f.Name == "Chart.lock":
			lock = new(chart.Lock)
			if err := yaml.Unmarshal(f.Data, &lock); err != nil {
				return nil, nil, errors.Wrap(err, "cannot load Chart.lock")
			}
			lockFile = f
			break
		case f.Name == "requirements.lock":
			lock = new(chart.Lock)
			if err := yaml.Unmarshal(f.Data, &lock); err != nil {
				return nil, nil, errors.Wrap(err, "cannot load requirements.lock")
			}
			lockFile = f
			break
		}
	}

	return lock, lockFile, nil
}

func GetPreparedChartDependenciesDir(ctx context.Context, lockDigest string, lockFileData []byte, chartFileData []byte, helmEnvSettings *cli.EnvSettings, buildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions) (string, error) {
	depsDir := GetChartDependenciesCacheDir(lockDigest)

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
				ChartExtender:               NewWerfChartStub(),
				SubchartExtenderFactoryFunc: nil,
			}

			if err := command_helpers.BuildChartDependenciesInDir(ctx, lockFileData, chartFileData, tmpDepsDir, helmEnvSettings, buildChartDependenciesOpts); err != nil {
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

func GiterministicFilesLoader(ctx context.Context, giterminismManager giterminism.Manager, loadDir string, helmEnvSettings *cli.EnvSettings, buildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions) ([]*chart.ChartExtenderBufferedFile, error) {
	gitFiles, err := giterminismManager.FileReader().LoadChartDir(ctx, loadDir)
	if err != nil {
		return nil, err
	}

	var chartFile *chart.ChartExtenderBufferedFile
	for _, f := range gitFiles {
		if f.Name == "Chart.yaml" {
			chartFile = f
			break
		}
	}

	res := gitFiles

	if chartFile != nil {
		if lock, lockFile, err := LoadLock(gitFiles); err != nil {
			return nil, fmt.Errorf("error loading chart lock file: %s", err)
		} else if lock == nil {
			if metadata, err := LoadMetadata(gitFiles); err != nil {
				return nil, fmt.Errorf("error loading chart metadata file: %s", err)
			} else if len(metadata.Dependencies) > 0 {
				logboek.Context(ctx).Error().LogLn("Cannot build chart dependencies and preload charts without lock file (.helm/Chart.lock or .helm/requirements.lock)")
				logboek.Context(ctx).Error().LogLn("It is recommended to add Chart.lock file to your project repository or remove chart dependencies.")
				logboek.Context(ctx).Error().LogLn()
				logboek.Context(ctx).Error().LogLn("To generate a lock file run 'werf helm dependency update .helm' and commit resulting .helm/Chart.lock (it is not required to commit whole .helm/charts directory).")
				logboek.Context(ctx).Error().LogLn()
			}
		} else {
			if depsDir, err := GetPreparedChartDependenciesDir(ctx, lock.Digest, lockFile.Data, chartFile.Data, helmEnvSettings, buildChartDependenciesOpts); err != nil {
				return nil, fmt.Errorf("")
			} else {
				localFiles, err := loader.GetFilesFromLocalFilesystem(depsDir)
				if err != nil {
					return nil, err
				}

				for _, f := range localFiles {
					if strings.HasPrefix(f.Name, "charts/") {
						f1 := new(chart.ChartExtenderBufferedFile)
						*f1 = chart.ChartExtenderBufferedFile(*f)
						res = append(res, f1)
						logboek.Context(ctx).Debug().LogF("-- GiterministicFilesLoader: loading subchart %q from dependencies dir %s\n", f.Name, depsDir)
					}
				}
			}
		}
	}

	return res, nil
}
