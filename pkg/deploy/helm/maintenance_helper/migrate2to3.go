package maintenance_helper

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"

	"github.com/werf/logboek"
)

func Migrate2To3(ctx context.Context, helm2ReleaseName, helm3ReleaseName, helm3Namespace string, maintenanceHelper *MaintenanceHelper) error {
	foundHelm3Release, err := maintenanceHelper.IsHelm3ReleaseExist(ctx, helm3ReleaseName)
	if err != nil {
		return fmt.Errorf("error checking existence of helm 3 release %q: %w", helm2ReleaseName, err)
	}

	if foundHelm3Release {
		return fmt.Errorf("found already existing helm 3 release %q", helm3ReleaseName)
	}

	foundHelm2Release, err := maintenanceHelper.IsHelm2ReleaseExist(ctx, helm2ReleaseName)
	if err != nil {
		return fmt.Errorf("error checking existence of helm 2 release %q: %w", helm2ReleaseName, err)
	}

	if !foundHelm2Release {
		return fmt.Errorf("not found helm 2 release %q", helm2ReleaseName)
	}

	releaseData, err := maintenanceHelper.GetHelm2ReleaseData(ctx, helm2ReleaseName)
	if err != nil {
		return fmt.Errorf("unable to get helm 2 release %q info: %w", helm2ReleaseName, err)
	}

	infos, err := maintenanceHelper.BuildHelm2ResourcesInfos(releaseData)
	if err != nil {
		return fmt.Errorf("error building resources infos for release %q: %w", helm2ReleaseName, err)
	}

	metadataAccessor := meta.NewAccessor()

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Migrating %d resources of the release %q", len(infos), helm2ReleaseName).DoError(func() error {
		for _, info := range infos {
			logboek.Context(ctx).Default().LogF("%s\n", info.ObjectName())

			helper := resource.NewHelper(info.Client, info.Mapping)

			obj, err := helper.Get(info.Namespace, info.Name)
			if apierrors.IsNotFound(err) {
				logboek.Context(ctx).Default().LogF("    %s not found: ignoring\n", info.ObjectName())
				continue
			} else if err != nil {
				return fmt.Errorf("error getting resource %s spec from %q namespace: %w", info.ObjectName(), info.Namespace, err)
			}

			annotations, err := metadataAccessor.Annotations(obj)
			if err != nil {
				return fmt.Errorf("error accessing annotations of %s: %w", info.ObjectName(), err)
			}
			if annotations == nil {
				annotations = make(map[string]string)
			}
			annotations["meta.helm.sh/release-name"] = helm3ReleaseName
			annotations["meta.helm.sh/release-namespace"] = helm3Namespace
			if err := metadataAccessor.SetAnnotations(obj, annotations); err != nil {
				return fmt.Errorf("error setting annotations of %s: %w", info.ObjectName(), err)
			}

			labels, err := metadataAccessor.Labels(obj)
			if err != nil {
				return fmt.Errorf("error accessing labels of %s: %w", info.ObjectName(), err)
			}
			if labels == nil {
				labels = make(map[string]string)
			}
			labels["app.kubernetes.io/managed-by"] = "Helm"
			if err := metadataAccessor.SetLabels(obj, labels); err != nil {
				return fmt.Errorf("error setting labels of %s: %w", info.ObjectName(), err)
			}

			if _, err := helper.Replace(info.Namespace, info.Name, false, obj); err != nil {
				return fmt.Errorf("error replacing %s: %w", info.ObjectName(), err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Creating helm 3 release %q", helm3ReleaseName).DoError(func() error {
		if err := maintenanceHelper.CreateHelm3ReleaseMetadataFromHelm2Release(ctx, helm3ReleaseName, helm3Namespace, releaseData); err != nil {
			return fmt.Errorf("unable to create helm 3 release %q: %w", helm3ReleaseName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).LogOptionalLn()
	if err := logboek.Context(ctx).Default().LogProcess("Deleting helm 2 metadata for release %q", helm2ReleaseName).DoError(func() error {
		if err := maintenanceHelper.DeleteHelm2ReleaseMetadata(ctx, helm2ReleaseName); err != nil {
			return fmt.Errorf("unable to delete helm 2 release storage metadata for the release %q: %w", helm2ReleaseName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	logboek.Context(ctx).LogOptionalLn()
	logboek.Context(ctx).Default().LogFDetails(" + Successfully migrated helm 2 release %q into helm 3 release %q in the %q namespace\n", helm2ReleaseName, helm3ReleaseName, helm3Namespace)

	return nil
}
