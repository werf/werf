package common

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/deploy_v2/maintenance_helper"

	"github.com/werf/werf/pkg/deploy_v2/helm_v3"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
)

func NewActionConfig(ctx context.Context, namespace string, commonCmdData *CmdData) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	*cmd_helm.Settings.GetNamespaceP() = namespace

	if err := helm_v3.InitActionConfig(ctx, cmd_helm.Settings, actionConfig, helm_v3.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
	}); err != nil {
		return nil, err
	}

	return actionConfig, nil
}

func Helm3ReleaseExistanceGuard(ctx context.Context, releaseName, namespace string, maintenanceHelper *maintenance_helper.MaintenanceHelper) error {
	list, err := maintenanceHelper.GetHelm3ReleasesList(ctx)
	if err != nil {
		return fmt.Errorf("error getting helm 3 releases list: %s", err)
	}

	for _, existingReleaseName := range list {
		if existingReleaseName == releaseName {
			return fmt.Errorf(`found existing helm 3 release %q in the namespace %q: cannot continue deploy process

Please use werf v1.2 to converge your application.`, releaseName, namespace)
		}
	}
	return nil
}

func CreateMaintenanceHelper(ctx context.Context, cmdData *CmdData, actionConfig *action.Configuration, kubeConfigOptions kube.KubeConfigOptions) (*maintenance_helper.MaintenanceHelper, error) {
	maintenanceOpts := maintenance_helper.MaintenanceHelperOptions{
		KubeConfigOptions: kubeConfigOptions,
	}

	if helmReleaseStorageType, err := GetHelmReleaseStorageType(*cmdData.HelmReleaseStorageType); err != nil {
		return nil, err
	} else {
		maintenanceOpts.Helm2ReleaseStorageType = helmReleaseStorageType
	}
	maintenanceOpts.Helm2ReleaseStorageNamespace = *cmdData.HelmReleaseStorageNamespace

	return maintenance_helper.NewMaintenanceHelper(actionConfig, maintenanceOpts), nil
}
