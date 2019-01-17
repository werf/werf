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

	m, err := getSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	repo := "REPO"
	tag := "DOCKER_TAG"
	namespace := "NAMESPACE"

	var images []DimgInfoGetter
	for _, dimg := range werfConfig.Dimgs {
		d := &DimgInfo{Config: dimg, WithoutRegistry: true, Repo: repo, Tag: tag}
		images = append(images, d)
	}

	serviceValues, err := GetServiceValues(werfConfig.Meta.Project, repo, namespace, tag, nil, images, ServiceValuesOptions{ForceBranch: "GIT_BRANCH"})

	werfChart, err := getWerfChart(projectDir, m, opts.Values, opts.SecretValues, opts.Set, opts.SetString, serviceValues)
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
