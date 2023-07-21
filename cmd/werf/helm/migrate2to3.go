package helm

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/maintenance_helper"
	"github.com/werf/werf/pkg/werf"
)

var migrate2To3CommonCmdData common.CmdData

var migrate2ToCmdData struct {
	Release         string
	TargetRelease   string
	TargetNamespace string

	Helm2ReleaseStorageNamespace string
	Helm2ReleaseStorageType      string
}

func setupHelm2ReleaseStorageNamespace(cmd *cobra.Command) {
	defaultValues := []string{
		os.Getenv("WERF_HELM2_RELEASE_STORAGE_NAMESPACE"),
		os.Getenv("WERF_HELM_RELEASE_STORAGE_NAMESPACE"),
		os.Getenv("TILLER_NAMESPACE"),
	}

	var defaultValue string
	for _, value := range defaultValues {
		if value != "" {
			defaultValue = value
			break
		}
	}

	cmd.Flags().StringVarP(&migrate2ToCmdData.Helm2ReleaseStorageNamespace, "helm2-release-storage-namespace", "", defaultValue, fmt.Sprintf("Helm 2 release storage namespace (same as --tiller-namespace for regular helm 2, defaults to $WERF_HELM2_RELEASE_STORAGE_NAMESPACE, or $WERF_HELM_RELEASE_STORAGE_NAMESPACE, or $TILLER_NAMESPACE, or \"kube-system\")"))
}

func setupHelm2ReleaseStorageType(cmd *cobra.Command) {
	defaultValues := []string{
		os.Getenv("WERF_HELM2_RELEASE_STORAGE_TYPE"),
		os.Getenv("WERF_HELM_RELEASE_STORAGE_TYPE"),
	}

	var defaultValue string
	for _, value := range defaultValues {
		if value != "" {
			defaultValue = value
			break
		}
	}

	cmd.Flags().StringVarP(&migrate2ToCmdData.Helm2ReleaseStorageType, "helm2-release-storage-type", "", defaultValue, fmt.Sprintf("Helm w storage driver to use. One of %[1]q or %[2]q, defaults to $WERF_HELM2_RELEASE_STORAGE_TYPE, or $WERF_HELM_RELEASE_STORAGE_TYPE, or %[1]q)", "configmap", "secret"))
}

func NewMigrate2To3Cmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &migrate2To3CommonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "migrate2to3",
		DisableFlagsInUseLine: true,
		Short:                 "Start a migration of your existing Helm 2 release to Helm 3.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&migrate2To3CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runMigrate2To3(ctx)
		},
	})

	common.SetupTmpDir(&migrate2To3CommonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&migrate2To3CommonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupKubeConfig(&migrate2To3CommonCmdData, cmd)
	common.SetupKubeConfigBase64(&migrate2To3CommonCmdData, cmd)
	common.SetupKubeContext(&migrate2To3CommonCmdData, cmd)

	common.SetupInsecureHelmDependencies(&migrate2To3CommonCmdData, cmd)

	common.SetupLogOptions(&migrate2To3CommonCmdData, cmd)

	cmd.Flags().StringVarP(&migrate2ToCmdData.Release, "release", "", os.Getenv("WERF_RELEASE"), "Existing helm 2 release name which should be migrated to helm 3 (default $WERF_RELEASE). Option also sets target name for a new helm 3 release, use --target-release option (or $WERF_TARGET_RELEASE) to specify a different helm 3 release name.")
	cmd.Flags().StringVarP(&migrate2ToCmdData.TargetRelease, "target-release", "", os.Getenv("WERF_TARGET_RELEASE"), "Target helm 3 release name (optional, default $WERF_TARGET_RELEASE, or the value of --release option, or $WERF_RELEASE)")
	cmd.Flags().StringVarP(&migrate2ToCmdData.TargetNamespace, "target-namespace", "", os.Getenv("WERF_NAMESPACE"), "Target kubernetes namespace for a new helm 3 release (default $WERF_NAMESPACE)")

	setupHelm2ReleaseStorageNamespace(cmd)
	setupHelm2ReleaseStorageType(cmd)

	return cmd
}

func runMigrate2To3(ctx context.Context) error {
	if err := werf.Init(*migrate2To3CommonCmdData.TmpDir, *migrate2To3CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	common.SetupOndemandKubeInitializer(*migrate2To3CommonCmdData.KubeContext, *migrate2To3CommonCmdData.KubeConfig, *migrate2To3CommonCmdData.KubeConfigBase64, *migrate2To3CommonCmdData.KubeConfigPathMergeList)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	existingReleaseName := migrate2ToCmdData.Release
	if existingReleaseName == "" {
		return fmt.Errorf("--release (or WERF_RELEASE env var) required! This option specifies existing helm 2 release name which should be migrated to helm 3, option also sets target name for a new helm 3 release, use --target-release option (or $WERF_TARGET_RELEASE) to specify a different helm 3 release name.")
	}

	targetReleaseName := migrate2ToCmdData.TargetRelease
	if targetReleaseName == "" {
		targetReleaseName = existingReleaseName
	}

	targetNamespace := migrate2ToCmdData.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("--target-namespace (or WERF_TARGET_NAMESPACE env var) required! Please specify target namespace for a new helm 3 release explicitly (specify \"default\" for the default namespace).")
	}

	kubeConfigOptions := kube.KubeConfigOptions{
		Context:             *migrate2To3CommonCmdData.KubeContext,
		ConfigPath:          *migrate2To3CommonCmdData.KubeConfig,
		ConfigDataBase64:    *migrate2To3CommonCmdData.KubeConfigBase64,
		ConfigPathMergeList: *migrate2To3CommonCmdData.KubeConfigPathMergeList,
	}

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *migrate2To3CommonCmdData.DockerConfig, *migrate2To3CommonCmdData.InsecureHelmDependencies)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %w", err)
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, common.GetOndemandKubeInitializer(), targetNamespace, helm_v3.Settings, actionConfig, helm.InitActionConfigOptions{
		KubeConfigOptions: kubeConfigOptions,
		RegistryClient:    helmRegistryClient,
	}); err != nil {
		return err
	}

	maintenanceOpts := maintenance_helper.MaintenanceHelperOptions{
		Helm2ReleaseStorageNamespace: migrate2ToCmdData.Helm2ReleaseStorageNamespace,
		Helm2ReleaseStorageType:      migrate2ToCmdData.Helm2ReleaseStorageType,
		KubeConfigOptions:            kubeConfigOptions,
	}

	maintenanceHelper := maintenance_helper.NewMaintenanceHelper(actionConfig, maintenanceOpts)

	if available, err := maintenanceHelper.CheckHelm2StorageAvailable(ctx); err != nil {
		return err
	} else if !available {
		return fmt.Errorf("helm 2 release storage is not available")
	}
	logboek.Context(ctx).Default().LogFDetails(" + Helm 2 release storage is available\n")

	if err := maintenance_helper.Migrate2To3(ctx, existingReleaseName, targetReleaseName, targetNamespace, maintenanceHelper); err != nil {
		return err
	}

	logboek.Context(ctx).Default().LogOptionalLn()
	logboek.Context(ctx).Default().LogFDetails(`Migration to werf v1.2 is almost done, please run "werf converge" command to bring resources to the state which is described in the repository .helm/templates directory.

Make sure "werf converge" command uses release name %q and namespace %q!
`, targetReleaseName, targetNamespace)

	return nil
}
