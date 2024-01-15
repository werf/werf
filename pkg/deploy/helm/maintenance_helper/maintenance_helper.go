package maintenance_helper

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	helm2to3_v3 "github.com/helm/helm-2to3/pkg/v3"
	v3_action "helm.sh/helm/v3/pkg/action"
	v3_rspb "helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
	v2_rspb "k8s.io/helm/pkg/proto/hapi/release"
	v2_releaseutil "k8s.io/helm/pkg/releaseutil"
	v2_storage "k8s.io/helm/pkg/storage"
	v2_driver "k8s.io/helm/pkg/storage/driver"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
)

func IsReleaseNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

type Helm2ReleaseData struct {
	Release *v2_rspb.Release
}

type Helm3ReleaseData struct {
	Release *v3_rspb.Release
}

type MaintenanceHelperOptions struct {
	Helm2ReleaseStorageNamespace string
	Helm2ReleaseStorageType      string
	KubeConfigOptions            kube.KubeConfigOptions
}

func NewMaintenanceHelper(v3ActionConfig *v3_action.Configuration, opts MaintenanceHelperOptions) *MaintenanceHelper {
	releaseStorageType := opts.Helm2ReleaseStorageType
	if releaseStorageType == "" {
		releaseStorageType = "configmap"
	}

	releaseStorageNamespace := opts.Helm2ReleaseStorageNamespace
	if releaseStorageNamespace == "" {
		releaseStorageNamespace = "kube-system"
	}

	return &MaintenanceHelper{
		Helm2ReleaseStorageNamespace: releaseStorageNamespace,
		Helm2ReleaseStorageType:      releaseStorageType,
		KubeConfigOptions:            opts.KubeConfigOptions,
		v3ActionConfig:               v3ActionConfig,
	}
}

type MaintenanceHelper struct {
	KubeConfigOptions kube.KubeConfigOptions

	Helm2ReleaseStorageNamespace string
	Helm2ReleaseStorageType      string

	v2Storage      *v2_storage.Storage
	v3ActionConfig *v3_action.Configuration
}

func (helper *MaintenanceHelper) initHelm2Storage() (*v2_storage.Storage, error) {
	if helper.v2Storage != nil {
		return helper.v2Storage, nil
	}

	var drv v2_driver.Driver
	switch helper.Helm2ReleaseStorageType {
	case "configmap":
		drv = v2_driver.NewConfigMaps(kube.Client.CoreV1().ConfigMaps(helper.Helm2ReleaseStorageNamespace))
	case "secret":
		drv = v2_driver.NewSecrets(kube.Client.CoreV1().Secrets(helper.Helm2ReleaseStorageNamespace))
	default:
		return nil, fmt.Errorf("unknown helm 2 release v2Storage type %q", helper.Helm2ReleaseStorageType)
	}
	helper.v2Storage = v2_storage.Init(drv)

	return helper.v2Storage, nil
}

func (helper *MaintenanceHelper) getResourcesFactory() (util.Factory, error) {
	configGetter, err := kube.NewKubeConfigGetter(kube.KubeConfigGetterOptions{KubeConfigOptions: helper.KubeConfigOptions})
	if err != nil {
		return nil, fmt.Errorf("error creating kube config getter: %w", err)
	}
	return util.NewFactory(configGetter), nil
}

func (helper *MaintenanceHelper) CheckHelm3StorageAvailable(ctx context.Context) (bool, error) {
	var err error

	logboek.Context(ctx).Debug().LogProcess("Checking helm 3 storage availability using history command").Do(func() {
		_, err = helper.v3ActionConfig.Releases.History("no-such-release")
	})

	if IsReleaseNotFoundErr(err) || err == nil {
		logboek.Context(ctx).Info().LogFDetails("+ Helm 3 storage available\n")
		return true, nil
	}

	logboek.Context(ctx).Info().LogFDetails("- Helm 3 storage is not available: %s\n", err)
	return false, nil
}

func (helper *MaintenanceHelper) CheckHelm2StorageAvailable(ctx context.Context) (bool, error) {
	storage, err := helper.initHelm2Storage()
	if err != nil {
		return false, fmt.Errorf("error initializing helm 2 v2Storage: %w", err)
	}

	logboek.Context(ctx).Debug().LogProcess("Checking helm 2 storage availability using history command").Do(func() {
		_, err = storage.History("no-such-release")
	})

	if IsReleaseNotFoundErr(err) || err == nil {
		logboek.Context(ctx).Info().LogFDetails("+ Helm 2 storage available\n")
		return true, nil
	}

	logboek.Context(ctx).Info().LogFDetails("- Helm 2 storage is not available: %s\n", err)
	return false, nil
}

func (helper *MaintenanceHelper) IsHelm3ReleaseExist(ctx context.Context, releaseName string) (bool, error) {
	var err error

	logboek.Context(ctx).Debug().LogProcess("Getting helm 3 release %q history", releaseName).Do(func() {
		_, err = helper.v3ActionConfig.Releases.History(releaseName)
	})

	if IsReleaseNotFoundErr(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("error getting helm 3 release %q history: %w", releaseName, err)
	}

	return true, nil
}

func (helper *MaintenanceHelper) IsHelm2ReleaseExist(ctx context.Context, releaseName string) (bool, error) {
	storage, err := helper.initHelm2Storage()
	if err != nil {
		return false, err
	}

	logboek.Context(ctx).Debug().LogProcess("Getting helm 2 release %q history", releaseName).Do(func() {
		_, err = storage.History(releaseName)
	})

	if IsReleaseNotFoundErr(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("error getting helm 2 release %q history: %w", releaseName, err)
	}

	return true, nil
}

func (helper *MaintenanceHelper) CreateHelm3ReleaseMetadataFromHelm2Release(ctx context.Context, release, namespace string, releaseData *Helm2ReleaseData) error {
	rls, err := helm2to3_v3.CreateRelease(releaseData.Release)
	if err != nil {
		return fmt.Errorf("cannot create helm 3 release %q metadata from helm 2 release %q metadata: %w", release, releaseData.Release.Name, err)
	}

	rls.Namespace = namespace

	if err := helper.v3ActionConfig.Releases.Create(rls); err != nil {
		return fmt.Errorf("error saving helm 3 release %q into storage: %w", release, err)
	}

	return nil
}

func (helper *MaintenanceHelper) GetHelm2ReleaseData(ctx context.Context, releaseName string) (*Helm2ReleaseData, error) {
	storage, err := helper.initHelm2Storage()
	if err != nil {
		return nil, err
	}

	releases, err := storage.ListFilterAll(func(rel *v2_rspb.Release) bool {
		return rel.Name == releaseName
	})
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("release not found")
	}

	v2_releaseutil.Reverse(releases, v2_releaseutil.SortByRevision)

	return &Helm2ReleaseData{Release: releases[0]}, nil
}

func (helper *MaintenanceHelper) BuildHelm2ResourcesInfos(releaseData *Helm2ReleaseData) ([]*resource.Info, error) {
	manifestBuffer := bytes.NewBufferString(releaseData.Release.Manifest)

	factory, err := helper.getResourcesFactory()
	if err != nil {
		return nil, err
	}

	schema, err := factory.Validator(v1.FieldValidationIgnore)
	if err != nil {
		return nil, err
	}

	return newBuilder(factory, releaseData.Release.Namespace).
		Unstructured().
		Schema(schema).
		Stream(manifestBuffer, "").
		Do().Infos()
}

func (helper *MaintenanceHelper) DeleteHelm2ReleaseMetadata(ctx context.Context, releaseName string) error {
	storage, err := helper.initHelm2Storage()
	if err != nil {
		return err
	}

	releases, err := storage.ListFilterAll(func(rel *v2_rspb.Release) bool {
		return rel.Name == releaseName
	})
	if err != nil {
		return err
	}

	if len(releases) == 0 {
		return nil
	}

	for _, rel := range releases {
		if _, err := storage.Delete(rel.Name, rel.Version); err != nil {
			return fmt.Errorf("error deleting helm 2 release %q version %d: %w", rel.Name, rel.Version, err)
		}
	}

	return nil
}

func newBuilder(factory util.Factory, namespace string) *resource.Builder {
	return factory.NewBuilder().
		ContinueOnError().
		NamespaceParam(namespace).
		DefaultNamespace().
		Flatten()
}
