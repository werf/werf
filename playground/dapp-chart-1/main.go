package main

import (
	"fmt"

	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/secret"
)

func main() {
	werfChart, err := deploy.GenerateDappChart(".", &secret.BaseManager{})
	if err != nil {
		panic(err)
	}

	err = werfChart.SetSecretValuesFile("mypath.yaml", &secret.BaseManager{})
	if err != nil {
		panic(err)
	}

	err = werfChart.SetValues(map[string]interface{}{
		"custom": "values",
		"service_info": map[string]interface{}{
			"version":     "1.0.0",
			"project_url": "https://github.com/flant/werf",
			"a number":    123.456,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", werfChart)
}
