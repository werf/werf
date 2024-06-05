package render

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	helmstorage "helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/werf/nelm/pkg/chrttree"
	helmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/kubeclnt"
	"github.com/werf/nelm/pkg/resrc"
	"github.com/werf/nelm/pkg/resrcpatcher"
	"github.com/werf/nelm/pkg/resrcprocssr"
	"github.com/werf/nelm/pkg/rlshistor"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/v2/pkg/deploy/secrets_manager"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/util"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag          string
	BundleDir    string
	RenderOutput string
	Validate     bool
	IncludeCRDs  bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "render",
		Short:                 "Render Kubernetes manifests from bundle",
		Long:                  common.GetLongCommandDescription(`Take locally extracted bundle or download bundle from the specified container registry using specified version tag or version mask and render it as Kubernetes manifests.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runRender(ctx) })
		},
	})

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, false)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd, false)
	common.SetupSecretValues(&commonCmdData, cmd, false)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)
	commonCmdData.SetupDisableDefaultSecretValues(cmd)

	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupNamespace(&commonCmdData, cmd, false)

	common.SetupKubeVersion(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}

	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag,
		"Provide exact tag version or semver-based pattern, werf will render the latest version of the specified bundle ($WERF_TAG or latest by default)")
	cmd.Flags().BoolVarP(&cmdData.IncludeCRDs, "include-crds", "", util.GetBoolEnvironmentDefaultTrue("WERF_INCLUDE_CRDS"),
		"Include CRDs in the templated output (default $WERF_INCLUDE_CRDS)")
	cmd.Flags().StringVarP(&cmdData.RenderOutput, "output", "", os.Getenv("WERF_RENDER_OUTPUT"),
		"Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by default)")
	cmd.Flags().StringVarP(&cmdData.BundleDir, "bundle-dir", "b", os.Getenv(("WERF_BUNDLE_DIR")),
		"Get extracted bundle from directory instead of registry (default $WERF_BUNDLE_DIR)")

	cmd.Flags().BoolVarP(&cmdData.Validate, "validate", "", util.GetBoolEnvironmentDefaultFalse("WERF_VALIDATE"), "Validate your manifests against the Kubernetes cluster you are currently pointing at (default $WERF_VALIDATE)")

	return cmd
}

func runRender(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	var isLocal bool
	switch {
	case cmdData.BundleDir != "":
		if *commonCmdData.Repo.Address != "" {
			return fmt.Errorf("only one of --bundle-dir or --repo should be specified, but both provided")
		}

		isLocal = true
	case *commonCmdData.Repo.Address == storage.LocalStorageAddress:
		return fmt.Errorf("--repo %s is not allowed, specify remote storage address", storage.LocalStorageAddress)
	case *commonCmdData.Repo.Address != "":
		isLocal = false
	default:
		return fmt.Errorf("either --bundle-dir or --repo required")
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *commonCmdData.DockerConfig, *commonCmdData.InsecureHelmDependencies)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %w", err)
	}

	namespace := common.GetNamespace(&commonCmdData)
	releaseName := common.GetOptionalRelease(&commonCmdData)

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, nil, namespace, helm_v3.Settings, actionConfig, helm.InitActionConfigOptions{RegistryClient: helmRegistryClient}); err != nil {
		return err
	}

	var bundleDir string
	if isLocal {
		bundleDir = cmdData.BundleDir
	} else {
		registryMirrors, err := common.GetContainerRegistryMirror(ctx, &commonCmdData)
		if err != nil {
			return fmt.Errorf("get container registry mirrors: %w", err)
		}

		if err := common.DockerRegistryInit(ctx, &commonCmdData, registryMirrors); err != nil {
			return err
		}

		bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
		if err != nil {
			return err
		}

		repoAddress, err := commonCmdData.Repo.GetAddress()
		if err != nil {
			return err
		}

		bundleDir = filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
		defer os.RemoveAll(bundleDir)

		if err := bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), bundleDir, bundlesRegistryClient); err != nil {
			return fmt.Errorf("unable to pull bundle: %w", err)
		}
	}

	secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{DisableSecretsDecryption: *commonCmdData.IgnoreSecretKey})

	bundle, err := chart_extender.NewBundle(ctx, bundleDir, helm_v3.Settings, helmRegistryClient, secretsManager, chart_extender.BundleOptions{
		SecretValueFiles:                  common.GetSecretValues(&commonCmdData),
		BuildChartDependenciesOpts:        command_helpers.BuildChartDependenciesOptions{IgnoreInvalidAnnotationsAndLabels: false},
		IgnoreInvalidAnnotationsAndLabels: false,
		ExtraAnnotations:                  userExtraAnnotations,
		ExtraLabels:                       userExtraLabels,
	})
	if err != nil {
		return err
	}

	if vals, err := helpers.GetBundleServiceValues(ctx, helpers.ServiceValuesOptions{
		Env:                      *commonCmdData.Environment,
		Namespace:                namespace,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         *commonCmdData.DockerConfig,
	}); err != nil {
		return fmt.Errorf("error creating service values: %w", err)
	} else {
		bundle.SetServiceValues(vals)
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender: bundle,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender {
			return chart_extender.NewWerfSubchart(ctx, secretsManager, chart_extender.WerfSubchartOptions{
				DisableDefaultSecretValues: *commonCmdData.DisableDefaultSecretValues,
			})
		},
	}

	networkParallelism := common.GetNetworkParallelism(&commonCmdData)

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	for key, value := range bundle.ExtraAnnotationsAndLabelsPostRenderer.ExtraAnnotations {
		if strings.HasPrefix(key, "project.werf.io/") ||
			strings.Contains(key, "ci.werf.io/") ||
			key == "werf.io/release-channel" {
			serviceAnnotations[key] = value
		} else {
			extraAnnotations[key] = value
		}
	}

	serviceAnnotations["werf.io/version"] = werf.Version
	if *commonCmdData.Environment != "" {
		serviceAnnotations["project.werf.io/env"] = *commonCmdData.Environment
	}

	extraLabels := bundle.ExtraAnnotationsAndLabelsPostRenderer.ExtraLabels

	var clientFactory *kubeclnt.ClientFactory
	if cmdData.Validate {
		clientFactory, err = kubeclnt.NewClientFactory()
		if err != nil {
			return fmt.Errorf("error creating kube client factory: %w", err)
		}
	}

	var releaseNamespaceOptions resrc.ReleaseNamespaceOptions
	if cmdData.Validate {
		releaseNamespaceOptions.Mapper = clientFactory.Mapper()
	}

	releaseNamespace := resrc.NewReleaseNamespace(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": lo.WithoutEmpty([]string{namespace, helm_v3.Settings.Namespace()})[0],
			},
		},
	}, releaseNamespaceOptions)

	// FIXME(ilya-lesikov): there is more chartpath options, are they needed?
	chartPathOptions := action.ChartPathOptions{}
	chartPathOptions.SetRegistryClient(actionConfig.RegistryClient)

	if !cmdData.Validate {
		mem := driver.NewMemory()
		mem.SetNamespace(releaseNamespace.Name())
		actionConfig.Releases = helmstorage.Init(mem)

		actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: ioutil.Discard}
		actionConfig.Capabilities = chartutil.DefaultCapabilities.Copy()

		if *commonCmdData.KubeVersion != "" {
			if kubeVersion, err := chartutil.ParseKubeVersion(*commonCmdData.KubeVersion); err != nil {
				return fmt.Errorf("invalid kube version %q: %w", kubeVersion, err)
			} else {
				actionConfig.Capabilities.KubeVersion = *kubeVersion
			}
		}
	}

	var historyOptions rlshistor.HistoryOptions
	if cmdData.Validate {
		historyOptions.Mapper = clientFactory.Mapper()
		historyOptions.DiscoveryClient = clientFactory.Discovery()
	}

	history, err := rlshistor.NewHistory(releaseName, releaseNamespace.Name(), actionConfig.Releases, historyOptions)
	if err != nil {
		return fmt.Errorf("error constructing release history: %w", err)
	}

	prevRelease, prevReleaseFound, err := history.LastRelease()
	if err != nil {
		return fmt.Errorf("error getting last deployed release: %w", err)
	}

	_, prevDeployedReleaseFound, err := history.LastDeployedRelease()
	if err != nil {
		return fmt.Errorf("error getting last deployed release: %w", err)
	}

	var newRevision int
	if prevReleaseFound {
		newRevision = prevRelease.Revision() + 1
	} else {
		newRevision = 1
	}

	var deployType helmcommon.DeployType
	if prevReleaseFound && prevDeployedReleaseFound {
		deployType = helmcommon.DeployTypeUpgrade
	} else if prevReleaseFound {
		deployType = helmcommon.DeployTypeInstall
	} else {
		deployType = helmcommon.DeployTypeInitial
	}

	chartTreeOptions := chrttree.ChartTreeOptions{
		StringSetValues: common.GetSetString(&commonCmdData),
		SetValues:       common.GetSet(&commonCmdData),
		FileValues:      common.GetSetFile(&commonCmdData),
		ValuesFiles:     common.GetValues(&commonCmdData),
	}
	if cmdData.Validate {
		chartTreeOptions.Mapper = clientFactory.Mapper()
		chartTreeOptions.DiscoveryClient = clientFactory.Discovery()
	}

	chartTree, err := chrttree.NewChartTree(
		ctx,
		bundle.Dir,
		releaseName,
		releaseNamespace.Name(),
		newRevision,
		deployType,
		actionConfig,
		chartTreeOptions,
	)
	if err != nil {
		return fmt.Errorf("error constructing chart tree: %w", err)
	}

	var prevRelGeneralResources []*resrc.GeneralResource
	if prevReleaseFound {
		prevRelGeneralResources = prevRelease.GeneralResources()
	}

	resProcessorOptions := resrcprocssr.DeployableResourcesProcessorOptions{
		NetworkParallelism: networkParallelism,
		ReleasableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
			resrcpatcher.NewExtraMetadataPatcher(extraAnnotations, extraLabels),
		},
		ReleasableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
			resrcpatcher.NewExtraMetadataPatcher(extraAnnotations, extraLabels),
		},
		DeployableStandaloneCRDsPatchers: []resrcpatcher.ResourcePatcher{
			resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
		},
		DeployableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
			resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
		},
		DeployableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
			resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
		},
	}
	if cmdData.Validate {
		resProcessorOptions.KubeClient = clientFactory.KubeClient()
		resProcessorOptions.Mapper = clientFactory.Mapper()
		resProcessorOptions.DiscoveryClient = clientFactory.Discovery()
		resProcessorOptions.AllowClusterAccess = true
	}

	resProcessor := resrcprocssr.NewDeployableResourcesProcessor(
		deployType,
		releaseName,
		releaseNamespace,
		chartTree.StandaloneCRDs(),
		chartTree.HookResources(),
		chartTree.GeneralResources(),
		prevRelGeneralResources,
		resProcessorOptions,
	)

	if err := resProcessor.Process(ctx); err != nil {
		return fmt.Errorf("error processing deployable resources: %w", err)
	}

	var output io.Writer
	if cmdData.RenderOutput != "" {
		if f, err := os.Create(cmdData.RenderOutput); err != nil {
			return fmt.Errorf("unable to open file %q: %w", cmdData.RenderOutput, err)
		} else {
			defer f.Close()
			output = f
		}
	} else {
		output = os.Stdout
	}

	if cmdData.IncludeCRDs {
		crds := resProcessor.DeployableStandaloneCRDs()

		for _, res := range crds {
			if err := renderResource(res.Unstructured(), res.FilePath(), output); err != nil {
				return fmt.Errorf("error rendering CRD %q: %w", res.HumanID(), err)
			}
		}
	}

	hooks := resProcessor.DeployableHookResources()

	for _, res := range hooks {
		if err := renderResource(res.Unstructured(), res.FilePath(), output); err != nil {
			return fmt.Errorf("error rendering hook resource %q: %w", res.HumanID(), err)
		}
	}

	resources := resProcessor.DeployableGeneralResources()

	for _, res := range resources {
		if err := renderResource(res.Unstructured(), res.FilePath(), output); err != nil {
			return fmt.Errorf("error rendering general resource %q: %w", res.HumanID(), err)
		}
	}

	return nil
}

func renderResource(unstruct *unstructured.Unstructured, path string, output io.Writer) error {
	resourceJsonBytes, err := runtime.Encode(unstructured.UnstructuredJSONScheme, unstruct)
	if err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}

	resourceYamlBytes, err := yaml.JSONToYAML(resourceJsonBytes)
	if err != nil {
		return fmt.Errorf("marshalling to YAML failed: %w", err)
	}

	prefixBytes := []byte(fmt.Sprintf("---\n# Source: %s\n", path))

	if _, err := output.Write(append(prefixBytes, resourceYamlBytes...)); err != nil {
		return fmt.Errorf("writing to output failed: %w", err)
	}

	return nil
}
