package helm_v3

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/werf/logboek"

	helm_kube "helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/klog"
)

type InitOptions struct {
	Debug bool
}

func Init(opts InitOptions) error {
	gofs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(gofs)
	pflag.CommandLine.AddGoFlagSet(gofs)
	pflag.CommandLine.Set("logtostderr", "true")

	return nil
}

func withEnvs(envsMap map[string]string, do func()) {
	for k, v := range envsMap {
		oldV := os.Getenv(k)
		os.Setenv(k, v)
		defer func(resetValue string) { os.Setenv(k, resetValue) }(oldV)
	}
	do()
}

func NewEnvSettings(namespace string) (res *cli.EnvSettings) {
	withEnvs(map[string]string{
		"HELM_NAMESPACE":         namespace,
		"HELM_KUBECONTEXT":       "",
		"HELM_KUBETOKEN":         "",
		"HELM_KUBEAPISERVER":     "",
		"HELM_PLUGINS":           "",
		"HELM_REGISTRY_CONFIG":   "",
		"HELM_REPOSITORY_CONFIG": "",
		"HELM_REPOSITORY_CACHE":  "",
		"HELM_DEBUG":             "",
	}, func() {
		res = cli.New()
	})

	res.Debug = logboek.Debug().IsAccepted()

	return
}

type InitActionConfigOptions struct {
	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration
}

func NewActionConfig(envSettings *cli.EnvSettings, opts InitActionConfigOptions) *action.Configuration {
	actionConfig := &action.Configuration{}
	InitActionConfig(envSettings, actionConfig, opts)
	return actionConfig
}

func InitActionConfig(envSettings *cli.EnvSettings, actionConfig *action.Configuration, opts InitActionConfigOptions) {
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(envSettings.RESTClientGetter(), envSettings.Namespace(), helmDriver, logboek.Debug().LogF); err != nil {
		log.Fatal(err)
	}
	if helmDriver == "memory" {
		loadReleasesInMemory(envSettings, actionConfig)
	}

	kubeClient := actionConfig.KubeClient.(*helm_kube.Client)
	kubeClient.ResourcesWaiter = NewResourcesWaiter(kubeClient, time.Now(), opts.StatusProgressPeriod, opts.HooksStatusProgressPeriod)
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
