package deploy

import (
	"fmt"

	"github.com/flant/werf/pkg/tag_scheme"

	"github.com/flant/werf/pkg/config"
)

type LintOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
}

func RunLint(projectDir string, werfConfig *config.WerfConfig, opts LintOptions) error {
	if debug() {
		fmt.Printf("Lint options: %#v\n", opts)
	}

	m, err := GetSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	imagesRepo := "REPO"
	tag := "GIT_BRANCH"
	tagScheme := tag_scheme.GitBranchScheme
	namespace := "NAMESPACE"

	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, true)

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagScheme, images)
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}

	return werfChart.Lint()
}
