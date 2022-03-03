package command_helpers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"

	"github.com/werf/logboek"
)

type BuildChartDependenciesOptions struct {
	Keyring     string
	SkipUpdate  bool
	Verify      downloader.VerificationStrategy
	LoadOptions *loader.LoadOptions
}

func BuildChartDependenciesInDir(ctx context.Context, chartFile, chartLockFile *chart.ChartExtenderBufferedFile, targetDir string, helmEnvSettings *cli.EnvSettings, registryClient *registry.Client, opts BuildChartDependenciesOptions) error {
	logboek.Context(ctx).Debug().LogF("-- BuildChartDependenciesInDir\n")

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %s", targetDir, err)
	}

	files := []*chart.ChartExtenderBufferedFile{chartFile, chartLockFile}

	for _, file := range files {
		if file == nil {
			continue
		}

		path := filepath.Join(targetDir, file.Name)
		if err := ioutil.WriteFile(path, file.Data, 0o644); err != nil {
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
		RegistryClient:   registryClient,
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
