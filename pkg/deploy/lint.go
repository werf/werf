package deploy

import (
	"fmt"

	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tag_strategy"

	"github.com/flant/werf/pkg/config"
)

type LintOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
	Env          string
}

func RunLint(projectDir string, werfConfig *config.WerfConfig, opts LintOptions) error {
	if debug() {
		fmt.Fprintf(logger.GetOutStream(), "Lint options: %#v\n", opts)
	}

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	imagesRepo := "REPO"
	tag := "GIT_BRANCH"
	tagStrategy := tag_strategy.GitBranch
	namespace := "NAMESPACE"

	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, true)

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagStrategy, images, ServiceValuesOptions{Env: opts.Env})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}
	defer ReleaseTmpWerfChart(werfChart.ChartDir)

	return werfChart.Lint(helm.HelmChartValuesOptions{
		Set:       opts.Set,
		SetString: opts.SetString,
		Values:    opts.Values,
	})
}
