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

func ReplaceHelmLintDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmLintDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmListDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmListDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmPackageDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmPackageDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmPullDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmPullDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmRollbackDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmRollbackDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmStatusDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmStatusDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmUninstallDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmUninstallDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmUpgradeDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmUpgradeDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmVerifyDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmVerifyDocs().LongMD,
	}
	return cmd
}

func ReplaceHelmVersionDocs(cmd *cobra.Command) *cobra.Command {
	cmd.Annotations = map[string]string{
		common.DocsLongMD: GetHelmVersionDocs().LongMD,
	}
	return cmd
}
