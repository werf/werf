package main

import (
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/kubedog/pkg/kube"
)

func main() {
	if err := lock.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Lock init error: %s\n", err)
		os.Exit(1)
	}

	if err := kube.Init(kube.InitOptions{}); err != nil {
		fmt.Fprintf(os.Stderr, "Kube init error: %s\n", err)
		os.Exit(1)
	}

	err := deploy.DeployHelmChart(os.Args[1], os.Args[2], os.Args[3], deploy.HelmChartOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deploying helm chart: %s\n", err)
		os.Exit(1)
	}
}
