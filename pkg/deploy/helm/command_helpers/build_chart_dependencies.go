package command_helpers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/werf/logboek"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
)

type BuildChartDependenciesOptions struct {
	Keyring     string
	SkipUpdate  bool
	Verify      downloader.VerificationStrategy
	LoadOptions *loader.LoadOptions
}

func BuildChartDependenciesInDir(ctx context.Context, chartFile *chart.ChartExtenderBufferedFile, chartLockFile *chart.ChartExtenderBufferedFile, requirementsFile *chart.ChartExtenderBufferedFile, requirementsLockFile *chart.ChartExtenderBufferedFile, targetDir string, helmEnvSettings *cli.EnvSettings, registryClientHandle *helm_v3.RegistryClientHandle, opts BuildChartDependenciesOptions) error {
	logboek.Context(ctx).Debug().LogF("-- BuildChartDependenciesInDir\n")

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %s", targetDir, err)
	}

	files := []*chart.ChartExtenderBufferedFile{chartFile, chartLockFile, requirementsFile, requirementsLockFile}

	for _, file := range files {
		if file == nil {
			continue
		}

		path := filepath.Join(targetDir, file.Name)
		if err := ioutil.WriteFile(path, file.Data, 0644); err != nil {
			return fmt.Errorf("error writing %q: %s", path, err)
		}
	}

	man := &downloader.Manager{
		Out:        logboek.Context(ctx).OutStream(),
		ChartPath:  targetDir,
		Keyring:    opts.Keyring,
		SkipUpdate: opts.SkipUpdate,
		Verify:     opts.Verify,

		Getters:          getter.All(helmEnvSettings),
		RegistryClient:   registryClientHandle.RegistryClient,
		RepositoryConfig: helmEnvSettings.RepositoryConfig,
		RepositoryCache:  helmEnvSettings.RepositoryCache,
		Debug:            helmEnvSettings.Debug,
	}

	currentLoaderOptions := loader.GlobalLoadOptions
	loader.GlobalLoadOptions = opts.LoadOptions
	defer func() {
		loader.GlobalLoadOptions = currentLoaderOptions
	}()

	err := man.Build()

	if e, ok := err.(downloader.ErrRepoNotFound); ok {
		return fmt.Errorf("%s. Please add the missing repos via 'helm repo add'", e.Error())
	}

	return err
}
