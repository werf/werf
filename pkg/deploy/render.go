package deploy

import (
	"fmt"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/tag_scheme"
)

type RenderOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
}

func RunRender(projectDir string, werfConfig *config.WerfConfig, opts RenderOptions) error {
	if debug() {
		fmt.Printf("Render options: %#v\n", opts)
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

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.SecretValues, serviceValues)
	if err != nil {
		return err
	}

	data, err := werfChart.Render(namespace)
	if err != nil {
		return err
	}

	if data != "" {
		fmt.Println(data)
	}

	return nil
}
