package helm2

import (
	"bytes"
	"context"
	"fmt"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	relutil "k8s.io/helm/pkg/releaseutil"

	rspb "k8s.io/helm/pkg/proto/hapi/release"

	"github.com/werf/logboek"

	"github.com/werf/kubedog/pkg/kube"

	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
)

type ReleaseData struct {
	Release *rspb.Release
}

type MaintenanceHelperOptions struct {
	ReleaseStorageNamespace string
	ReleaseStorageType      string
	KubeConfigOptions       kube.KubeConfigOptions
}

func NewMaintenanceHelper(opts MaintenanceHelperOptions) *MaintenanceHelper {
	releaseStorageType := opts.ReleaseStorageType
	if releaseStorageType == "" {
		releaseStorageType = "configmap"
	}

	releaseStorageNamespace := opts.ReleaseStorageNamespace
	if releaseStorageNamespace == "" {
		releaseStorageNamespace = "kube-system"
	}

	return &MaintenanceHelper{
		ReleaseStorageNamespace: releaseStorageNamespace,
		ReleaseStorageType:      releaseStorageType,
		KubeConfigOptions:       opts.KubeConfigOptions,
	}
}

type MaintenanceHelper struct {
	ReleaseStorageNamespace string
	ReleaseStorageType      string
	KubeConfigOptions       kube.KubeConfigOptions

	storage *storage.Storage
}

func (helper *MaintenanceHelper) initStorage() (*storage.Storage, error) {
	if helper.storage != nil {
		return helper.storage, nil
	}

	var drv driver.Driver
	switch helper.ReleaseStorageType {
	case "configmap":
		drv = driver.NewConfigMaps(kube.Client.CoreV1().ConfigMaps(helper.ReleaseStorageNamespace))
	case "secret":
		drv = driver.NewSecrets(kube.Client.CoreV1().Secrets(helper.ReleaseStorageNamespace))
	default:
		return nil, fmt.Errorf("unknown helm release storage type %q", helper.ReleaseStorageType)
	}
	helper.storage = storage.Init(drv)

	return helper.storage, nil
}

func (helper *MaintenanceHelper) getResourcesFactory() (cmdutil.Factory, error) {
	configGetter, err := NewKubeConfigGetter("", helper.KubeConfigOptions)
	if err != nil {
		return nil, fmt.Errorf("error creating kube config getter: %s", err)
	}

	return cmdutil.NewFactory(configGetter), nil
}

func (helper *MaintenanceHelper) CheckStorageAvailable(ctx context.Context) (bool, error) {
	storage, err := helper.initStorage()
	if err != nil {
		return false, fmt.Errorf("error initializing helm 2 storage: %s", err)
	}

	_, err = storage.ListReleases()
	return err == nil, nil
}

func (helper *MaintenanceHelper) GetReleasesList(ctx context.Context) ([]string, error) {
	storage, err := helper.initStorage()
	if err != nil {
		return nil, err
	}

	releases, err := storage.ListReleases()
	if err != nil {
		return nil, err
	}

	var res []string
AppendUniqReleases:
	for _, rel := range releases {
		for _, name := range res {
			if name == rel.Name {
				continue AppendUniqReleases
			}
		}
		res = append(res, rel.Name)
	}

	logboek.Context(ctx).Debug().LogF("-- MaintenanceHelper GetReleasesList: %#v\n", res)

	return res, nil
}

func (helper *MaintenanceHelper) GetReleaseData(ctx context.Context, releaseName string) (*ReleaseData, error) {
	storage, err := helper.initStorage()
	if err != nil {
		return nil, err
	}

	releases, err := storage.ListFilterAll(func(rel *rspb.Release) bool {
		return rel.Name == releaseName
	})
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("release not found")
	}

	relutil.Reverse(releases, relutil.SortByRevision)

	return &ReleaseData{Release: releases[0]}, nil
}

func (helper *MaintenanceHelper) BuildResourcesInfos(releaseData *ReleaseData) ([]*resource.Info, error) {
	manifestBuffer := bytes.NewBufferString(releaseData.Release.Manifest)

	factory, err := helper.getResourcesFactory()
	if err != nil {
		return nil, err
	}

	return factory.NewBuilder().
		Unstructured().
		ContinueOnError().
		NamespaceParam(releaseData.Release.Namespace).
		DefaultNamespace().
		Stream(manifestBuffer, "").
		Flatten().
		Do().Infos()
}

func (helper *MaintenanceHelper) ForgetReleaseStorageMetadata(ctx context.Context, releaseName string) error {
	storage, err := helper.initStorage()
	if err != nil {
		return err
	}

	releases, err := storage.ListFilterAll(func(rel *rspb.Release) bool {
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
			return fmt.Errorf("error deleting release %q version %d: %s", rel.Name, rel.Version)
		}
	}

	return nil
}

func NewKubeConfigGetter(namespace string, kubeConfigOpts kube.KubeConfigOptions) (genericclioptions.RESTClientGetter, error) {
	var configGetter genericclioptions.RESTClientGetter

	if kubeConfigOpts.ConfigDataBase64 != "" {
		if getter, err := NewClientGetterFromConfigData(kubeConfigOpts.Context, kubeConfigOpts.ConfigDataBase64); err != nil {
			return nil, fmt.Errorf("unable to create kube client getter (context=%q, config-data-base64=%q): %s", kubeConfigOpts.Context, kubeConfigOpts.ConfigPath, err)
		} else {
			configGetter = getter
		}
	} else {
		configFlags := genericclioptions.NewConfigFlags(true)

		configFlags.Context = new(string)
		*configFlags.Context = kubeConfigOpts.Context

		configFlags.KubeConfig = new(string)
		*configFlags.KubeConfig = kubeConfigOpts.ConfigPath

		if namespace != "" {
			configFlags.Namespace = new(string)
			*configFlags.Namespace = namespace
		}

		configGetter = configFlags
	}

	return configGetter, nil
}
