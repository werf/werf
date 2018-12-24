package deploy

import (
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/config"
)

type RenderOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
}

func RunRender(projectName, projectDir string, dappfile []*config.Dimg, opts RenderOptions) error {
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
	for _, dimg := range dappfile {
		d := &DimgInfo{Config: dimg, WithoutRegistry: true, Repo: repo, Tag: tag}
		images = append(images, d)
	}

	serviceValues, err := GetServiceValues(projectName, repo, namespace, tag, nil, images, ServiceValuesOptions{ForceBranch: "GIT_BRANCH"})

	dappChart, err := getDappChart(projectDir, m, opts.Values, opts.SecretValues, opts.Set, opts.SetString, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(dappChart.ChartDir)
	}

	data, err := dappChart.Render(namespace)
	if err != nil {
		return err
	}

	fmt.Println(data)

	return nil
}
