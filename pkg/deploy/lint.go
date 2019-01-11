package deploy

import (
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/config"
)

type LintOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
}

func RunLint(projectName, projectDir string, dappfile *config.Dappfile, opts LintOptions) error {
	if debug() {
		fmt.Printf("Lint options: %#v\n", opts)
	}

	m, err := getSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	repo := "REPO"
	tag := "DOCKER_TAG"
	namespace := "NAMESPACE"

	var images []DimgInfoGetter
	for _, dimg := range dappfile.Dimgs {
		d := &DimgInfo{Config: dimg, WithoutRegistry: true, Repo: repo, Tag: tag}
		images = append(images, d)
	}

	serviceValues, err := GetServiceValues(projectName, repo, namespace, tag, nil, images, ServiceValuesOptions{ForceBranch: "GIT_BRANCH"})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	dappChart, err := getDappChart(projectDir, m, opts.Values, opts.SecretValues, opts.Set, opts.SetString, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(dappChart.ChartDir)
	}

	return dappChart.Lint()
}
