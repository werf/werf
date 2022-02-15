package render

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/bundles"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag          string
	BundleDir    string
	RenderOutput string
	Validate     bool
	IncludeCRDs  bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "render",
		Short:                 "Render Kubernetes manifests from bundle",
		Long:                  common.GetLongCommandDescription(`Take locally extracted bundle or download bundle from the specified container registry using specified version tag or version mask and render it as Kubernetes manifests.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.GetContext()

			defer global_warnings.PrintGlobalWarnings(common.GetContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runRender(ctx) })
		},
	}

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupFinalStagesStorageOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptionsDefaultQuiet(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}

	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag,
		"Provide exact tag version or semver-based pattern, werf will render the latest version of the specified bundle ($WERF_TAG or latest by default)")
	cmd.Flags().BoolVarP(&cmdData.IncludeCRDs, "include-crds", "", common.GetBoolEnvironmentDefaultTrue("WERF_INCLUDE_CRDS"),
		"Include CRDs in the templated output (default $WERF_INCLUDE_CRDS)")
	cmd.Flags().StringVarP(&cmdData.RenderOutput, "output", "", os.Getenv("WERF_RENDER_OUTPUT"),
		"Write render output to the specified file instead of stdout ($WERF_RENDER_OUTPUT by default)")
	cmd.Flags().StringVarP(&cmdData.BundleDir, "bundle-dir", "b", os.Getenv(("WERF_BUNDLE_DIR")),
		"Get extracted bundle from directory instead of registry (default $WERF_BUNDLE_DIR)")

	cmd.Flags().BoolVarP(&cmdData.Validate, "validate", "", common.GetBoolEnvironmentDefaultFalse("WERF_VALIDATE"), "Validate your manifests against the Kubernetes cluster you are currently pointing at (default $WERF_VALIDATE)")

	return cmd
}

func runRender(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	var isLocal bool
	switch {
	case cmdData.BundleDir != "":
		if *commonCmdData.StagesStorage != "" {
			return fmt.Errorf("only one of --bundle-dir or --repo should be specified, but both provided")
		}
		if *commonCmdData.FinalStagesStorage != "" {
			return fmt.Errorf("only one of --bundle-dir or --final-repo should be specified, but both provided")
		}

		isLocal = true
	case *commonCmdData.StagesStorage == storage.LocalStorageAddress:
		return fmt.Errorf("--repo %s is not allowed, specify remote storage address", storage.LocalStorageAddress)
	case *commonCmdData.StagesStorage != "":
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

	helmRegistryClientHandle, err := common.NewHelmRegistryClientHandle(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %s", err)
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, nil, *commonCmdData.Namespace, helm_v3.Settings, helmRegistryClientHandle, actionConfig, helm.InitActionConfigOptions{}); err != nil {
		return err
	}

	var bundleDir string
	if isLocal {
		bundleDir = cmdData.BundleDir
	} else {
		if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
			return err
		}

		bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
		if err != nil {
			return err
		}

		repoAddress, err := common.GetStagesStorageAddress(&commonCmdData)
		if err != nil {
			return err
		}

		bundleDir = filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewV4().String())
		defer os.RemoveAll(bundleDir)

		if err := bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), bundleDir, bundlesRegistryClient); err != nil {
			return fmt.Errorf("unable to pull bundle: %s", err)
		}
	}

	namespace := common.GetNamespace(&commonCmdData)
	releaseName := common.GetOptionalRelease(&commonCmdData)

	if *commonCmdData.Environment != "" {
		userExtraAnnotations["project.werf.io/env"] = *commonCmdData.Environment
	}

	bundle, err := chart_extender.NewBundle(ctx, bundleDir, helm_v3.Settings, helmRegistryClientHandle, chart_extender.BundleOptions{
		ExtraAnnotations: userExtraAnnotations,
		ExtraLabels:      userExtraLabels,
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
		return fmt.Errorf("error creating service values: %s", err)
	} else {
		bundle.SetServiceValues(vals)
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender:               bundle,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender { return chart_extender.NewWerfSubchart() },
	}

	var output io.Writer
	if cmdData.RenderOutput != "" {
		if f, err := os.Create(cmdData.RenderOutput); err != nil {
			return fmt.Errorf("unable to open file %q: %s", cmdData.RenderOutput, err)
		} else {
			defer f.Close()
			output = f
		}
	} else {
		output = os.Stdout
	}

	helmTemplateCmd, _ := helm_v3.NewTemplateCmd(actionConfig, output, helm_v3.TemplateCmdOptions{
		ChainPostRenderer: bundle.ChainPostRenderer,
		ValueOpts: &values.Options{
			ValueFiles:   common.GetValues(&commonCmdData),
			StringValues: common.GetSetString(&commonCmdData),
			Values:       common.GetSet(&commonCmdData),
			FileValues:   common.GetSetFile(&commonCmdData),
		},
		Validate:    &cmdData.Validate,
		IncludeCrds: &cmdData.IncludeCRDs,
	})

	if err := helmTemplateCmd.RunE(helmTemplateCmd, []string{releaseName, bundleDir}); err != nil {
		return fmt.Errorf("helm templates rendering failed: %s", err)
	}

	return nil
}
