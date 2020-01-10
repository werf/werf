package rollback

import (
	"fmt"
	"path/filepath"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CommonCmdData common.CmdData

var CmdData struct {
	helm.RollbackOptions
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "rollback RELEASE_NAME REVISION",
		Short:                 "Rollback a release to the specified revision",
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	common.SetupKubeConfig(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)
	common.SetupHelmReleaseStorageNamespace(&CommonCmdData, cmd)
	common.SetupHelmReleaseStorageType(&CommonCmdData, cmd)
	common.SetupReleasesHistoryMax(&CommonCmdData, cmd)

	cmd.Flags().BoolVar(&CmdData.DisableHooks, "no-hooks", false, "Prevent hooks from running during rollback")
	cmd.Flags().BoolVar(&CmdData.Recreate, "recreate-pods", false, "Perform pods restart for the resource if applicable")
	cmd.Flags().BoolVar(&CmdData.Wait, "wait", false, "If set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	cmd.Flags().BoolVar(&CmdData.Force, "force", false, "Force resource update through delete/recreate if needed")
	cmd.Flags().BoolVar(&CmdData.CleanupOnFail, "cleanup-on-fail", false, "Allow deletion of new resources created in this rollback when rollback failed")
	cmd.Flags().Int64Var(&CmdData.Timeout, "timeout", 300, "Time in seconds to wait for any individual Kubernetes operation (like Jobs for hooks)")

	return cmd
}

func runRollback(releaseName string, revision int32) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	helmReleaseStorageType, err := common.GetHelmReleaseStorageType(*CommonCmdData.HelmReleaseStorageType)
	if err != nil {
		return err
	}

	deployInitOptions := deploy.InitOptions{
		HelmInitOptions: helm.InitOptions{
			KubeConfig:                  *CommonCmdData.KubeConfig,
			KubeContext:                 *CommonCmdData.KubeContext,
			HelmReleaseStorageNamespace: *CommonCmdData.HelmReleaseStorageNamespace,
			HelmReleaseStorageType:      helmReleaseStorageType,
			ReleasesMaxHistory:          *CommonCmdData.ReleasesHistoryMax,
		},
	}
	if err := deploy.Init(deployInitOptions); err != nil {
		return err
	}

	if err := kube.Init(kube.InitOptions{KubeContext: *CommonCmdData.KubeContext, KubeConfig: *CommonCmdData.KubeConfig}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	common.LogKubeContext(kube.Context)

	if err := common.InitKubedog(); err != nil {
		return fmt.Errorf("cannot init kubedog: %s", err)
	}

	if err := helm.Rollback(releaseName, revision, CmdData.RollbackOptions); err != nil {
		return err
	}

	return nil
}
