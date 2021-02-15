package maintenance_helper

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
)

func Migrate2To3(ctx context.Context, helm2ReleaseName, helm3ReleaseName, helm3Namespace string, maintenanceHelper *MaintenanceHelper) error {
	existingHelm3Releases, err := maintenanceHelper.GetHelm3ReleasesList(ctx)
	if err != nil {
		return fmt.Errorf("error getting existing helm 3 releases to perform check: %s", err)
	}

	foundHelm3Release := false
	for _, releaseName := range existingHelm3Releases {
		if releaseName == helm3ReleaseName {
			foundHelm3Release = true
			break
		}
	}

	if foundHelm3Release {
		return fmt.Errorf("found already existing helm 3 release %q", helm3ReleaseName)
	}

	existingReleases, err := maintenanceHelper.GetHelm2ReleasesList(ctx)
	if err != nil {
		return fmt.Errorf("error getting existing helm 2 releases to perform check: %s", err)
	}

	foundHelm2Release := false
	for _, releaseName := range existingReleases {
		if releaseName == helm2ReleaseName {
			foundHelm2Release = true
			break
		}
	}

	if !foundHelm2Release {
		return fmt.Errorf("not found helm 2 release %q", helm2ReleaseName)
	}

	releaseData, err := maintenanceHelper.GetHelm2ReleaseData(ctx, helm2ReleaseName)
	if err != nil {
		return fmt.Errorf("unable to get helm 2 release %q info: %s", helm2ReleaseName, err)
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Creating helm 3 release %q", helm3ReleaseName).DoError(func() error {
		if err := maintenanceHelper.CreateHelm3ReleaseMetadataFromHelm2Release(ctx, helm3ReleaseName, helm3Namespace, releaseData); err != nil {
			return fmt.Errorf("unable to create helm 3 release %q: %s", helm3ReleaseName)
		}
		return nil
	}); err != nil {
		return err
	}

	infos, err := maintenanceHelper.BuildHelm2ResourcesInfos(releaseData)
	if err != nil {
		return fmt.Errorf("error building resources infos for release %q: %s", helm2ReleaseName, err)
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Migrating %d resources of the release %q", len(infos), helm2ReleaseName).DoError(func() error {
		for _, info := range infos {
			logboek.Context(ctx).Default().LogF("%s\n", info.ObjectName())

			helper := resource.NewHelper(info.Client, info.Mapping)

			if _, err := helper.Patch(info.Namespace, info.Name, types.StrategicMergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"app.kubernetes.io/managed-by":"Helm"},"annotations":{"meta.helm.sh/release-name":%q,"meta.helm.sh/release-namespace":%q}}}`, helm3ReleaseName, helm3Namespace)), nil); err != nil {
				return fmt.Errorf("error patching %s: %s", info.ObjectName(), err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Deleting helm 2 metadata for release %q", helm2ReleaseName).DoError(func() error {
		if err := maintenanceHelper.DeleteHelm2ReleaseMetadata(ctx, helm2ReleaseName); err != nil {
			return fmt.Errorf("unable to delete helm 2 release storage metadata for the release %q: %s", helm2ReleaseName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).LogOptionalLn()
	logboek.Context(ctx).Default().LogFDetails(" + Successfully migrated helm 2 release %q into helm 3 release %q in the %q namespace\n", helm2ReleaseName, helm3ReleaseName, helm3Namespace)

	return nil
}
