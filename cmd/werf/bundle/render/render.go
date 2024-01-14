package render

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/bundles"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/util"
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
	common.SetupSecretValues(&commonCmdData, cmd)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)
	commonCmdData.SetupDisableDefaultSecretValues(cmd)

	common.SetupRelease(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd)

	common.SetupKubeVersion(&commonCmdData, cmd)

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
		if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
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

	if *commonCmdData.Environment != "" {
		userExtraAnnotations["project.werf.io/env"] = *commonCmdData.Environment
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

	helmTemplateCmd, _ := helm_v3.NewTemplateCmd(actionConfig, output, helm_v3.TemplateCmdOptions{
		StagesSplitter:    helm.NewStagesSplitter(),
		ChainPostRenderer: bundle.ChainPostRenderer,
		ValueOpts: &values.Options{
			ValueFiles:   common.GetValues(&commonCmdData),
			StringValues: common.GetSetString(&commonCmdData),
			Values:       common.GetSet(&commonCmdData),
			FileValues:   common.GetSetFile(&commonCmdData),
		},
		Validate:    &cmdData.Validate,
		IncludeCrds: &cmdData.IncludeCRDs,
		KubeVersion: commonCmdData.KubeVersion,
	})

	if err := helmTemplateCmd.RunE(helmTemplateCmd, []string{releaseName, bundleDir}); err != nil {
		return fmt.Errorf("helm templates rendering failed: %w", err)
	}

	return nil
}
