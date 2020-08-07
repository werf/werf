package common

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

func SetupCmdActionConfig(cmd *cobra.Command, actionConfig *action.Configuration) {
	oldRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := actionConfig.Init(helm_v3.Settings.RESTClientGetter(), helm_v3.Settings.Namespace(), helmDriver, helm_v3.Debug); err != nil {
			log.Fatal(err)
		}
		if helmDriver == "memory" {
			helm_v3.LoadReleasesInMemory(actionConfig)
		}

		if oldRunE != nil {
			return oldRunE(cmd, args)
		}
		return nil
	}

	cmd.Flags().StringVarP(helm_v3.Settings.GetNamespaceP(), "namespace", "n", *helm_v3.Settings.GetNamespaceP(), "namespace scope for this request")
}
