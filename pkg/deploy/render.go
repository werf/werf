package deploy

import "fmt"

type RenderOptions struct {
	ProjectDir   string
	Values       []string
	SecretValues []string
	Set          []string
}

func RunRender(opts RenderOptions) error {
	if debug() {
		fmt.Printf("Render options: %#v\n", opts)
	}

	s, err := getOptionalSecret(opts.ProjectDir, opts.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	serviceValues, err := GetServiceValues("PROJECT_NAME", "REPO", "NAMESPACE", "DOCKER_TAG", nil, nil, ServiceValuesOptions{
		Fake:            true,
		WithoutRegistry: true,
	})

	dappChart, err := getDappChart(opts.ProjectDir, s, opts.Values, opts.SecretValues, opts.Set, serviceValues)
	if err != nil {
		return err
	}

	data, err := dappChart.Render()
	if err != nil {
		return err
	}

	fmt.Println(data)

	return nil
}
