package main

import (
	"fmt"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/deploy/secret"
)

func main() {
	dappChart, err := deploy.GenerateDappChart(".", &secret.BaseManager{})
	if err != nil {
		panic(err)
	}

	err = dappChart.SetSecretValuesFile("mypath.yaml", &secret.BaseManager{})
	if err != nil {
		panic(err)
	}

	err = dappChart.SetValues(map[string]interface{}{
		"custom": "values",
		"service_info": map[string]interface{}{
			"version":     "1.0.0",
			"project_url": "https://github.com/flant/dapp",
			"a number":    123.456,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", dappChart)
}
