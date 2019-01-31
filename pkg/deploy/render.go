package deploy

import (
	"fmt"
	"os"

	"github.com/flant/werf/pkg/config"
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
	tag := "DOCKER_TAG"
	namespace := "NAMESPACE"

	images := GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, true)

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, nil, images, ServiceValuesOptions{ForceBranch: "GIT_BRANCH"})

	werfChart, err := PrepareWerfChart(GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, opts.Values, opts.SecretValues, opts.Set, opts.SetString, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(werfChart.ChartDir)
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
