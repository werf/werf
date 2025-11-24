package copy

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/3p-helm/pkg/werf/helmopts"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	From string
	To   string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "copy",
		Short:                 "Copy published bundle into another location",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag and copy it either into a different tag within the same container registry or into another container registry.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runCopy(ctx) })
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repos")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.StubSetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)

	commonCmdData.SetupHelmCompatibleChart(cmd, true)
	commonCmdData.SetupRenameChart(cmd)

	cmd.Flags().StringVarP(&cmdData.From, "from", "", os.Getenv("WERF_FROM"), "Source address of the bundle to copy, specify bundle archive using schema `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema `[docker://]REPO:TAG` or without schema.")
	cmd.Flags().StringVarP(&cmdData.To, "to", "", os.Getenv("WERF_TO"), "Destination address of the bundle to copy, specify bundle archive using schema `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema `[docker://]REPO:TAG` or without schema.")

	return cmd
}

func runCopy(ctx context.Context) error {
	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                &commonCmdData,
		InitDockerRegistry: true,
		InitWerf:           true,
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	if cmdData.From == "" {
		return fmt.Errorf("--from=ADDRESS param required")
	}

	fromAddrRaw := cmdData.From

	toAddrRaw := cmdData.To
	if toAddrRaw == "" {
		return fmt.Errorf("--to=ADDRESS param required")
	}

	fromAddr, err := ref.ParseAddr(fromAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid from addr %q: %w", fromAddrRaw, err)
	}

	toAddr, err := ref.ParseAddr(toAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid to addr %q: %w", toAddrRaw, err)
	}

	var fromRegistry, toRegistry docker_registry.Interface

	if fromAddr.RegistryAddress != nil {
		fromRegistry, err = common.CreateDockerRegistry(ctx, fromAddr.RegistryAddress.Repo, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
		if err != nil {
			return err
		}
	}

	if toAddr.RegistryAddress != nil {
		toRegistry, err = common.CreateDockerRegistry(ctx, toAddr.RegistryAddress.Repo, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
		if err != nil {
			return err
		}
	}

	if commonCmdData.HelmCompatibleChart && commonCmdData.RenameChart != "" {
		return fmt.Errorf("incompatible options specified, could not use --helm-compatible-chart and --rename-chart=%q at the same time", commonCmdData.RenameChart)
	}

	return logboek.Context(ctx).LogProcess("Copy bundle").DoError(func() error {
		logboek.Context(ctx).LogFDetails("From: %s\n", fromAddr.String())
		logboek.Context(ctx).LogFDetails("To: %s\n", toAddr.String())

		return bundles.Copy(ctx, fromAddr, toAddr, bundles.CopyOptions{
			BundlesRegistryClient: bundlesRegistryClient,
			FromRegistryClient:    fromRegistry,
			ToRegistryClient:      toRegistry,
			HelmCompatibleChart:   commonCmdData.HelmCompatibleChart,
			RenameChart:           commonCmdData.RenameChart,
			HelmOptions: helmopts.HelmOptions{
				ChartLoadOpts: helmopts.ChartLoadOptions{
					ChartType: helmopts.ChartTypeBundle,
					NoSecrets: true,
				},
			},
		})
	})
}
