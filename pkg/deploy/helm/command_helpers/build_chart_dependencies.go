package command_helpers

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"

	"github.com/werf/logboek"
)

type BuildChartDependenciesOptions struct {
	Keyring                           string
	SkipUpdate                        bool
	Verify                            downloader.VerificationStrategy
	LoadOptions                       *loader.LoadOptions
	IgnoreInvalidAnnotationsAndLabels bool
}

func BuildChartDependenciesInDir(ctx context.Context, targetDir string, helmEnvSettings *cli.EnvSettings, registryClient *registry.Client, opts BuildChartDependenciesOptions) error {
	logboek.Context(ctx).Debug().LogF("-- BuildChartDependenciesInDir\n")

	man := &downloader.Manager{
		Out:               logboek.Context(ctx).OutStream(),
		ChartPath:         targetDir,
		Keyring:           opts.Keyring,
		SkipUpdate:        opts.SkipUpdate,
		Verify:            opts.Verify,
		AllowMissingRepos: true,

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
		return fmt.Errorf("%w. Please add the missing repos via 'helm repo add'", e)
	}

	return err
}
