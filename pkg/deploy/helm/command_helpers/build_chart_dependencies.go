package command_helpers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

func BuildChartDependenciesInDir(ctx context.Context, lockFileData []byte, chartFileData []byte, targetDir string, helmEnvSettings *cli.EnvSettings, opts BuildChartDependenciesOptions) error {
	logboek.Context(ctx).Debug().LogF("-- BuildChartDependenciesInDir\n")

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %s", targetDir, err)
	}

	lockFilePath := filepath.Join(targetDir, "Chart.lock")
	if err := ioutil.WriteFile(lockFilePath, lockFileData, 0644); err != nil {
		return fmt.Errorf("error writing %q: %s", lockFilePath)
	}

	chartFilePath := filepath.Join(targetDir, "Chart.yaml")
	if err := ioutil.WriteFile(chartFilePath, chartFileData, 0644); err != nil {
		return fmt.Errorf("error writing %q: %s", chartFilePath)
	}

	man := &downloader.Manager{
		Out:        logboek.ProxyOutStream(),
		ChartPath:  targetDir,
		Keyring:    opts.Keyring,
		SkipUpdate: opts.SkipUpdate,
		Verify:     opts.Verify,

		Getters:          getter.All(helmEnvSettings),
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

	return nil
}
