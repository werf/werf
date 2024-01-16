package helm

import (
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/werf/cmd/werf/common"
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

func ReplaceHelmDependencyDocs(cmd *cobra.Command) *cobra.Command {
	for i, c := range cmd.Commands() {
		switch c.Use {
		case "build CHART":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmDependencyBuildDocs().LongMD,
			}
		case "update CHART":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmDependencyUpdateDocs().LongMD,
			}
		case "list CHART":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmDependencyListDocs().LongMD,
			}
		}
	}
	return cmd
}

func ReplaceHelmGetDocs(cmd *cobra.Command) *cobra.Command {
	for i, c := range cmd.Commands() {
		switch c.Use {
		case "hooks RELEASE_NAME":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmGetHooksDocs().LongMD,
			}
		case "all RELEASE_NAME":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmGetAllDocs().LongMD,
			}
		case "values RELEASE_NAME":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmGetValuesDocs().LongMD,
			}
		case "manifest RELEASE_NAME":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmGetManifestDocs().LongMD,
			}
		case "notes RELEASE_NAME":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmGetNotesDocs().LongMD,
			}
		case "metadata RELEASE_NAME":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmGetMetadataDocs().LongMD,
			}
		}
	}
	return cmd
}

func ReplaceHelmPluginDocs(cmd *cobra.Command) *cobra.Command {
	for i, c := range cmd.Commands() {
		switch c.Use {
		case "list":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmPluginListDocs().LongMD,
			}
		case "uninstall <plugin>...":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmPluginUninstallDocs().LongMD,
			}
		case "update <plugin>...":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmPluginUpdateDocs().LongMD,
			}
		case "install [options] <path|url>...":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmPluginInstallDocs().LongMD,
			}
		}
	}
	return cmd
}

func ReplaceHelmRepoDocs(cmd *cobra.Command) *cobra.Command {
	for i, c := range cmd.Commands() {
		switch c.Use {
		case "add [NAME] [URL]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmRepoAddDocs().LongMD,
			}
		case "index [DIR]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmRepoIndexDocs().LongMD,
			}
		case "list":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmRepoListDocs().LongMD,
			}
		case "remove [REPO1 [REPO2 ...]]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmRepoRemoveDocs().LongMD,
			}
		case "update [REPO1 [REPO2 ...]]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmRepoUpdateDocs().LongMD,
			}
		}
	}
	return cmd
}

func ReplaceHelmSearchDocs(cmd *cobra.Command) *cobra.Command {
	for i, c := range cmd.Commands() {
		switch c.Use {
		case "hub [KEYWORD]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmSearchHubDocs().LongMD,
			}
		case "repo [keyword]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmSearchRepoDocs().LongMD,
			}
		}
	}
	return cmd
}

func ReplaceHelmShowDocs(cmd *cobra.Command) *cobra.Command {
	for i, c := range cmd.Commands() {
		switch c.Use {
		case "all [CHART]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmShowAllDocs().LongMD,
			}
		case "chart [CHART]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmShowChartDocs().LongMD,
			}
		case "crds [CHART]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmShowCRDsDocs().LongMD,
			}
		case "readme [CHART]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmShowReadmeDocs().LongMD,
			}
		case "values [CHART]":
			cmd.Commands()[i].Annotations = map[string]string{
				common.DocsLongMD: GetHelmShowValuesDocs().LongMD,
			}
		}
	}
	return cmd
}
