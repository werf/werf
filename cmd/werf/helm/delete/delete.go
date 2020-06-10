package delete

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

var cmdData struct {
	helm.DeleteOptions
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "delete RELEASE_NAME",
		Short:                 "Delete release from Kubernetes with all resources associated with the last release revision",
		Aliases:               []string{"del", "remove", "rm"},
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := common.ValidateMinimumNArgs(1, args, cmd); err != nil {
				return err
			}
			return runDelete(args)
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().BoolVar(&cmdData.DisableHooks, "no-hooks", false, "Prevent hooks from running during deletion")
	cmd.Flags().BoolVar(&cmdData.Purge, "purge", false, "Remove the release from the store and make its name free for later use")
	cmd.Flags().Int64Var(&cmdData.Timeout, "timeout", 300, "Time in seconds to wait for any individual Kubernetes operation (like Jobs for hooks)")

	return cmd
}

func runDelete(releaseNames []string) error {
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

	errors := []string{}
	for _, releaseName := range releaseNames {
		if err := helm.Delete(releaseName, cmdData.DeleteOptions); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("following errors have occured during removal of specified releases: %s", strings.Join(errors, "; "))
	}

	return nil
}
