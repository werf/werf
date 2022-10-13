package helm

import (
	"github.com/spf13/cobra"
	"github.com/werf/werf/cmd/werf/common"
	"helm.sh/helm/v3/pkg/action"
)

func ReplaceHelmCreateDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmCreateDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmEnvDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmEnvDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmHistoryDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmHistoryDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmInstallDocs(cmd *cobra.Command, client *action.Install) (*cobra.Command, *action.Install) {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmInstallDocs().LongMD,
	}
	return cmd, client
}
