package rollback

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	helm.RollbackOptions
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "rollback RELEASE_NAME REVISION",
		Short:                 "Rollback a release to the specified revision",
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := common.ValidateArgumentCount(2, args, cmd); err != nil {
				return err
			}

			revision, err := common.ConvertInt32Value(args[1])
			if err != nil {
				return err
			}

			return runRollback(args[0], revision)
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&commonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().BoolVar(&cmdData.DisableHooks, "no-hooks", false, "Prevent hooks from running during rollback")
	cmd.Flags().BoolVar(&cmdData.Recreate, "recreate-pods", false, "Perform pods restart for the resource if applicable")
	cmd.Flags().BoolVar(&cmdData.Wait, "wait", false, "If set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	cmd.Flags().BoolVar(&cmdData.Force, "force", false, "Force resource update through delete/recreate if needed")
	cmd.Flags().BoolVar(&cmdData.CleanupOnFail, "cleanup-on-fail", false, "Allow deletion of new resources created in this rollback when rollback failed")
	cmd.Flags().Int64Var(&cmdData.Timeout, "timeout", 300, "Time in seconds to wait for any individual Kubernetes operation (like Jobs for hooks)")

	return cmd
}

func runRollback(releaseName string, revision int32) error {
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
			ReleasesMaxHistory:          *commonCmdData.ReleasesHistoryMax,
		},
	}
	if err := deploy.Init(deployInitOptions); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *commonCmdData.KubeContext, KubeConfig: *commonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	common.LogKubeContext(kube.Context)

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	if err := helm.Rollback(releaseName, revision, cmdData.RollbackOptions); err != nil {
		return err
	}

	return nil
}
