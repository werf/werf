package helm

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/werf/3p-helm/pkg/action"
	"github.com/werf/3p-helm/pkg/cli"
	helm_kube "github.com/werf/3p-helm/pkg/kube"
	kubefake "github.com/werf/3p-helm/pkg/kube/fake"
	"github.com/werf/3p-helm/pkg/registry"
	"github.com/werf/3p-helm/pkg/release"
	"github.com/werf/3p-helm/pkg/storage/driver"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
)

type InitActionConfigOptions struct {
	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration
	KubeConfigOptions         kube.KubeConfigOptions
	ReleasesHistoryMax        int
	RegistryClient            *registry.Client
	KubeToken                 string
	KubeAPIServerName         string
	KubeCAPath                string
	KubeTLSServerName         string
	KubeSkipTLSVerify         bool
	QPSLimit                  int
	BurstLimit                int
}

func InitActionConfig(ctx context.Context, kubeInitializer KubeInitializer, namespace string, envSettings *cli.EnvSettings, actionConfig *action.Configuration, opts InitActionConfigOptions) error {
	configGetter, err := kube.NewKubeConfigGetter(kube.KubeConfigGetterOptions{
		KubeConfigOptions: opts.KubeConfigOptions,
		Namespace:         namespace,
		BearerToken:       opts.KubeToken,
		APIServer:         opts.KubeAPIServerName,
		CAFile:            opts.KubeCAPath,
		TLSServerName:     opts.KubeTLSServerName,
		SkipTLSVerify:     opts.KubeSkipTLSVerify,
		QPSLimit:          opts.QPSLimit,
		BurstLimit:        opts.BurstLimit,
	})
	if err != nil {
		return fmt.Errorf("error creating kube config getter: %w", err)
	}

	*envSettings.GetConfigP() = configGetter
	*envSettings.GetNamespaceP() = namespace
	if opts.KubeConfigOptions.Context != "" {
		envSettings.KubeContext = opts.KubeConfigOptions.Context
	}
	if opts.KubeConfigOptions.ConfigPath != "" {
		envSettings.KubeConfig = opts.KubeConfigOptions.ConfigPath
	}
	if opts.ReleasesHistoryMax != 0 {
		envSettings.MaxHistory = opts.ReleasesHistoryMax
	}

	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(envSettings.RESTClientGetter(), envSettings.Namespace(), helmDriver, logboek.Context(ctx).Debug().LogF); err != nil {
		return fmt.Errorf("action config init failed: %w", err)
	}
	if helmDriver == "memory" {
		loadReleasesInMemory(envSettings, actionConfig)
	}

	kubeClient := actionConfig.KubeClient.(*helm_kube.Client)
	kubeClient.Namespace = namespace
	kubeClient.ResourcesWaiter = NewResourcesWaiter(kubeInitializer, kubeClient, time.Now(), opts.StatusProgressPeriod, opts.HooksStatusProgressPeriod)
	kubeClient.Extender = NewHelmKubeClientExtender()

	actionConfig.Log = func(f string, a ...interface{}) {
		logboek.Context(ctx).Info().LogFDetails(fmt.Sprintf("%s\n", f), a...)
	}

	if opts.RegistryClient != nil {
		actionConfig.RegistryClient = opts.RegistryClient
	}

	return nil
}

// This function loads releases into the memory storage if the
// environment variable is properly set.
func loadReleasesInMemory(envSettings *cli.EnvSettings, actionConfig *action.Configuration) {
	filePaths := strings.Split(os.Getenv("HELM_MEMORY_DRIVER_DATA"), ":")
	if len(filePaths) == 0 {
		return
	}

	store := actionConfig.Releases
	mem, ok := store.Driver.(*driver.Memory)
	if !ok {
		// For an unexpected reason we are not dealing with the memory storage driver.
		return
	}

	actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}

	for _, path := range filePaths {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("Unable to read memory driver data", err)
		}

		releases := []*release.Release{}
		if err := yaml.Unmarshal(b, &releases); err != nil {
			log.Fatal("Unable to unmarshal memory driver data: ", err)
		}

		for _, rel := range releases {
			if err := store.Create(rel); err != nil {
				log.Fatal(err)
			}
		}
	}
	// Must reset namespace to the proper one
	mem.SetNamespace(envSettings.Namespace())
}
