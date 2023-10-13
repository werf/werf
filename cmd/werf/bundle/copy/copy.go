package copy

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/bundles"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var cmdData struct {
	// TODO(2.0): Legacy {
	Repo  string
	Tag   string
	ToTag string
	// TODO(2.0): } Legacy

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
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)

	commonCmdData.SetupHelmCompatibleChart(cmd, true)
	commonCmdData.SetupRenameChart(cmd)

	cmd.Flags().StringVarP(&cmdData.Repo, "repo", "", os.Getenv("WERF_REPO"), "Deprecated param, use --from=ADDR instead. Source address of bundle which should be copied.")
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", os.Getenv("WERF_TAG"), "Deprecated param, use --from=REPO:TAG instead. Provide from tag version of the bundle to copy ($WERF_TAG or latest by default).")
	cmd.Flags().StringVarP(&cmdData.ToTag, "to-tag", "", os.Getenv("WERF_TO_TAG"), "Deprecated param, use --to=REPO:TAG instead. Provide to tag version of the bundle to copy ($WERF_TO_TAG or same as --tag by default).")

	cmd.Flags().StringVarP(&cmdData.From, "from", "", os.Getenv("WERF_FROM"), "Source address of the bundle to copy, specify bundle archive using schema `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema `[docker://]REPO:TAG` or without schema.")
	cmd.Flags().StringVarP(&cmdData.To, "to", "", os.Getenv("WERF_TO"), "Destination address of the bundle to copy, specify bundle archive using schema `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema `[docker://]REPO:TAG` or without schema.")

	return cmd
}

func runCopy(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	var fromAddrRaw string
	if cmdData.From == "" && cmdData.Repo == "" {
		return fmt.Errorf("--from=ADDRESS param required")
	} else if cmdData.From != "" {
		fromAddrRaw = cmdData.From
	} else {
		logboek.Context(ctx).Warn().LogF("Please use --from=ADDRESS param instead of deprecated --repo=ADDRESS param\n")
		fromAddrRaw = cmdData.Repo
	}

	toAddrRaw := cmdData.To
	if toAddrRaw == "" {
		return fmt.Errorf("--to=ADDRESS param required")
	}

	fromAddr, err := bundles.ParseAddr(fromAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid from addr %q: %w", fromAddrRaw, err)
	}

	toAddr, err := bundles.ParseAddr(toAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid to addr %q: %w", toAddrRaw, err)
	}

	var fromRegistry, toRegistry docker_registry.Interface

	if fromAddr.RegistryAddress != nil {
		// TODO(2.0): remove legacy compatibility param
		if cmdData.Tag != "" {
			logboek.Context(ctx).Warn().LogF("Please use --from=REPO:TAG tag specification instead of deprecated --tag=TAG param\n")
			fromAddr.RegistryAddress.Tag = cmdData.Tag
		}

		fromRegistry, err = common.CreateDockerRegistry(fromAddr.RegistryAddress.Repo, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
		if err != nil {
			return err
		}
	}

	if toAddr.RegistryAddress != nil {
		// TODO(2.0): remove legacy compatibility param
		if cmdData.ToTag != "" {
			logboek.Context(ctx).Warn().LogF("Please use --to=REPO:TAG tag specification instead of deprecated --to-tag=TAG param\n")
			toAddr.RegistryAddress.Tag = cmdData.ToTag
		}

		toRegistry, err = common.CreateDockerRegistry(toAddr.RegistryAddress.Repo, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
		if err != nil {
			return err
		}
	}

	if *commonCmdData.HelmCompatibleChart && *commonCmdData.RenameChart != "" {
		return fmt.Errorf("incompatible options specified, could not use --helm-compatible-chart and --rename-chart=%q at the same time", *commonCmdData.RenameChart)
	}

	return logboek.Context(ctx).LogProcess("Copy bundle").DoError(func() error {
		logboek.Context(ctx).LogFDetails("From: %s\n", fromAddr.String())
		logboek.Context(ctx).LogFDetails("To: %s\n", toAddr.String())

		return bundles.Copy(ctx, fromAddr, toAddr, bundles.CopyOptions{
			BundlesRegistryClient: bundlesRegistryClient,
			FromRegistryClient:    fromRegistry,
			ToRegistryClient:      toRegistry,
			HelmCompatibleChart:   *commonCmdData.HelmCompatibleChart,
			RenameChart:           *commonCmdData.RenameChart,
		})
	})
}
