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
}

func RunLint(projectName, projectDir string, dappfile []*config.Dimg, opts LintOptions) error {
	if debug() {
		fmt.Printf("Lint options: %#v\n", opts)
	}

	m, err := getSafeSecretManager(projectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	images := []DimgInfoGetter{}
	for _, dimg := range dappfile {
		d := &DimgInfo{Config: dimg, WithoutRegistry: true}
		images = append(images, d)
	}

	serviceValues, err := GetServiceValues(projectName, "REPO", "NAMESPACE", "TAG", nil, images, ServiceValuesOptions{})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	dappChart, err := getDappChart(projectDir, m, opts.Values, opts.SecretValues, opts.Set, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(dappChart.ChartDir)
	}

	return dappChart.Lint()
}
