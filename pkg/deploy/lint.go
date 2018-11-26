package deploy

import (
	"fmt"
	"os"
)

type LintOptions struct {
	ProjectDir   string
	Values       []string
	SecretValues []string
	Set          []string

	// TODO: remove this after full port to go
	Dimgs []*DimgInfoGetterStub
}

func RunLint(opts LintOptions) error {
	if debug() {
		fmt.Printf("Lint options: %#v\n", opts)
	}

	m, err := getSafeSecretManager(opts.ProjectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	images := []DimgInfoGetter{}
	for _, dimg := range opts.Dimgs {
		if debug() {
			fmt.Printf("DimgInfoGetterStub: %#v\n", dimg)
		}
		images = append(images, dimg)
	}

	serviceValues, err := GetServiceValues("PROJECT_NAME", "REPO", "NAMESPACE", "DOCKER_TAG", nil, images, ServiceValuesOptions{
		Fake:            true,
		WithoutRegistry: true,
	})

	dappChart, err := getDappChart(opts.ProjectDir, m, opts.Values, opts.SecretValues, opts.Set, serviceValues)
	if err != nil {
		return err
	}
	if !debug() {
		// Do not remove tmp chart in debug
		defer os.RemoveAll(dappChart.ChartDir)
	}

	return dappChart.Lint()
}
