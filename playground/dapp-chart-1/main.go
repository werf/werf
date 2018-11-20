package main

import (
	"fmt"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/secret"
)

func main() {
	dappChart, err := deploy.GenerateDappChart(".", deploy.DappChartOptions{
		Secret:       &secret.AesSecret{},
		SecretValues: []string{"mypath.yaml"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", dappChart)
}
