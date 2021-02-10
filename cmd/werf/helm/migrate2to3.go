package helm

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/deploy/helm"
	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/kubedog/pkg/kube"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
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

func NewMigrate2To3Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "migrate2to3",
		DisableFlagsInUseLine: true,
		Short:                 "Start a migration of your existing helm 2 release to helm 3",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&migrate2To3CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runMigrate2To3(common.BackgroundContext())
		},
	}

	common.SetupTmpDir(&migrate2To3CommonCmdData, cmd)
	common.SetupHomeDir(&migrate2To3CommonCmdData, cmd)

	common.SetupKubeConfig(&migrate2To3CommonCmdData, cmd)
	common.SetupKubeConfigBase64(&migrate2To3CommonCmdData, cmd)
	common.SetupKubeContext(&migrate2To3CommonCmdData, cmd)

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
		return fmt.Errorf("initialization error: %s", err)
	}

	common.SetupOndemandKubeInitializer(*migrate2To3CommonCmdData.KubeContext, *migrate2To3CommonCmdData.KubeConfig, *migrate2To3CommonCmdData.KubeConfigBase64)
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
		Context:          *migrate2To3CommonCmdData.KubeContext,
		ConfigPath:       *migrate2To3CommonCmdData.KubeConfig,
		ConfigDataBase64: *migrate2To3CommonCmdData.KubeConfigBase64,
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, common.GetOndemandKubeInitializer(), targetNamespace, cmd_helm.Settings, actionConfig, helm.InitActionConfigOptions{KubeConfigOptions: kubeConfigOptions}); err != nil {
		return err
	}

	maintenanceOpts := maintenance_helper.MaintenanceHelperOptions{
		Helm2ReleaseStorageNamespace: migrate2ToCmdData.Helm2ReleaseStorageNamespace,
		Helm2ReleaseStorageType:      migrate2ToCmdData.Helm2ReleaseStorageType,
		KubeConfigOptions:            kubeConfigOptions,
	}

	maintenanceHelper := maintenance_helper.NewMaintenanceHelper(actionConfig, maintenanceOpts)

	existingHelm3Releases, err := maintenanceHelper.GetHelm3ReleasesList(ctx)
	if err != nil {
		return fmt.Errorf("error getting existing helm 3 releases to perform check: %s", err)
	}

	foundHelm3Release := false
	for _, releaseName := range existingHelm3Releases {
		if releaseName == targetReleaseName {
			foundHelm3Release = true
			break
		}
	}

	if foundHelm3Release {
		return fmt.Errorf("found already existing helm 3 release %q", targetReleaseName)
	}

	if available, err := maintenanceHelper.CheckHelm2StorageAvailable(ctx); err != nil {
		return err
	} else if !available {
		return fmt.Errorf("helm 2 release storage is not available")
	}

	logboek.Context(ctx).Default().LogFDetails(" + Helm 2 release storage is available\n")

	existingReleases, err := maintenanceHelper.GetHelm2ReleasesList(ctx)
	if err != nil {
		return fmt.Errorf("error getting existing helm 2 releases to perform check: %s", err)
	}

	foundHelm2Release := false
	for _, releaseName := range existingReleases {
		if releaseName == existingReleaseName {
			foundHelm2Release = true
			break
		}
	}

	if !foundHelm2Release {
		return fmt.Errorf("not found helm 2 release %q", existingReleaseName)
	}

	logboek.Context(ctx).Default().LogFDetails(" + Found helm 2 release %q\n", existingReleaseName)

	releaseData, err := maintenanceHelper.GetHelm2ReleaseData(ctx, existingReleaseName)
	if err != nil {
		return fmt.Errorf("unable to get helm 2 release %q info: %s", existingReleaseName, err)
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Creating helm 3 release %q", targetReleaseName).DoError(func() error {
		if err := maintenanceHelper.CreateHelm3ReleaseMetadataFromHelm2Release(ctx, targetReleaseName, targetNamespace, releaseData); err != nil {
			return fmt.Errorf("unable to create helm 3 release %q: %s", targetReleaseName)
		}
		return nil
	}); err != nil {
		return err
	}

	infos, err := maintenanceHelper.BuildHelm2ResourcesInfos(releaseData)
	if err != nil {
		return fmt.Errorf("error building resources infos for release %q: %s", existingReleaseName, err)
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Migrating %d resources of the release %q", len(infos), existingReleaseName).DoError(func() error {
		for _, info := range infos {
			logboek.Context(ctx).Default().LogF("%s\n", info.ObjectName())

			helper := resource.NewHelper(info.Client, info.Mapping)

			if _, err := helper.Patch(info.Namespace, info.Name, types.StrategicMergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"app.kubernetes.io/managed-by":"Helm"},"annotations":{"meta.helm.sh/release-name":%q,"meta.helm.sh/release-namespace":%q}}}`, targetReleaseName, targetNamespace)), nil); err != nil {
				return fmt.Errorf("error patching %s: %s", info.ObjectName(), err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Deleting helm 2 metadata for release %q", existingReleaseName).DoError(func() error {
		if err := maintenanceHelper.DeleteHelm2ReleaseMetadata(ctx, existingReleaseName); err != nil {
			return fmt.Errorf("unable to delete helm 2 release storage metadata for the release %q: %s", existingReleaseName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).Default().LogFDetails(`Migration to helm 3 is almost done.

Please run "werf converge" command to perform adoption of existing resources into a newly created helm 3 release and bring resources to the state which is described in the repository .helm/templates directory.

Make sure "werf converge" command uses release name %q and namespace %q!
`, targetReleaseName, targetNamespace)

	return nil
}
