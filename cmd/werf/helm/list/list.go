package ls

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

var CmdData struct {
	helm.LsOptions
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "list [FILTER]",
		Short:                 "List werf releases",
		Aliases:               []string{"ls"},
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := common.ValidateMaximumNArgs(1, args, cmd); err != nil {
				return err
			}

			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}
			return runLs(filter)
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&CmdData.Short, "short", "q", false, "Output short listing format")
	cmd.Flags().StringVarP(&CmdData.Offset, "offset", "o", "", "Next release name in the list, used to offset from start value")
	cmd.Flags().BoolVarP(&CmdData.Reverse, "reverse", "r", false, "Reverse the sort order (descending by default)")
	cmd.Flags().Int64VarP(&CmdData.Max, "max", "m", 256, "Maximum number of releases to fetch")
	cmd.Flags().BoolVarP(&CmdData.All, "all", "a", false, "Show all releases, not just the ones marked DEPLOYED")
	cmd.Flags().BoolVar(&CmdData.Deleted, "deleted", false, "Show deleted releases")
	cmd.Flags().BoolVar(&CmdData.Deleting, "deleting", false, "Show releases that are currently being deleted")
	cmd.Flags().BoolVar(&CmdData.Pending, "pending", false, "Show pending releases")
	cmd.Flags().StringVar(&CmdData.Namespace, "namespace", "", "Show releases within a specific namespace")
	cmd.Flags().UintVar(&CmdData.ColWidth, "col-width", 60, "Specifies the max column width of output")
	cmd.Flags().StringVar(&CmdData.OutputFormat, "output", "table", "Output the specified format (json, yaml or table)")

	return cmd
}

func runLs(filter string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream(), LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	helmReleaseStorageType, err := common.GetHelmReleaseStorageType(*commonCmdData.HelmReleaseStorageType)
	if err != nil {
		return err
	}

	deployInitOptions := deploy.InitOptions{
		HelmInitOptions: helm.InitOptions{
			KubeConfig:                  *commonCmdData.KubeConfig,
			KubeContext:                 *commonCmdData.KubeContext,
			HelmReleaseStorageNamespace: *commonCmdData.HelmReleaseStorageNamespace,
			HelmReleaseStorageType:      helmReleaseStorageType,
			ReleasesMaxHistory:          0,
		},
	}
	if err := deploy.Init(deployInitOptions); err != nil {
		return err
	}

	if err := helm.Ls(os.Stdout, filter, CmdData.LsOptions); err != nil {
		return err
	}

	return nil
}
