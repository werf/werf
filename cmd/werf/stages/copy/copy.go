package copy

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	helm_v3 "github.com/werf/3p-helm-for-werf-helm/cmd/helm"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build/stages"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
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
		Short:                 "Copy stages between container registry and archive storage",
		Example:               "",
		Long:                  common.GetLongCommandDescription(GetCopyDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
			common.DocsLongMD: GetCopyDocs().LongMD,
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

// TODO: IMPLEMENT THIS ONE
func runCopy(ctx context.Context) error {
	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		//TODO: is needed?
		InitDockerRegistry: true,
		//TODO: is needed?
		InitWerf: true,
		//TODO: is needed?
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	//TODO: is needed?
	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	if cmdData.From == "" {
		return fmt.Errorf("--from=ADDRESS param required")
	}

	if cmdData.To == "" {
		return fmt.Errorf("--to=ADDRESS param required")
	}

	fromAddrRaw := cmdData.From
	toAddrRaw := cmdData.To

	//TODO подумай что с этим можно сделать - не нравится вызов из bundles
	fromAddr, err := bundles.ParseAddr(fromAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid from add %q: %w", fromAddrRaw, err)
	}

	//TODO подумай что с этим можно сделать - не нравится вызов из bundles
	toAddr, err := bundles.ParseAddr(toAddrRaw)
	if err != nil {
		return fmt.Errorf("invalid to addr %q: %w", toAddrRaw, err)
	}

	return logboek.Context(ctx).LogProcess("Copy stages").DoError(func() error {
		logboek.Context(ctx).Info().LogFDetails("From: %s\n", fromAddr.String())
		logboek.Context(ctx).Info().LogFDetails("To: %s\n", toAddr.String())

		return stages.Copy(ctx, fromAddr, toAddr, stages.CopyOptions{})
	})
}
