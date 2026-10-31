package copy

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/cmd/werf/common"
	"github.com/werf/werf/v3/pkg/deploy/bundles"
	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/tmp_manager"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/pkg/werf/global_warnings"
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
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.From, "from", "", os.Getenv("WERF_FROM"), "Source address of the bundle to copy, specify bundle archive (an OCI image layout, or an archive written by werf 3.12 or older) using schema `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema `[docker://]REPO:TAG` or without schema.")
	cmd.Flags().StringVarP(&cmdData.To, "to", "", os.Getenv("WERF_TO"), "Destination address of the bundle to copy, specify bundle archive, written as an OCI image layout, using schema `archive:PATH_TO_ARCHIVE.tar.gz`, specify remote bundle with schema `[docker://]REPO[:TAG]` or without schema. A remote bundle gets its chart version from TAG and is pushed under that version and under TAG; without TAG it keeps the chart version and the tag of the source bundle, or, copied from an archive written by werf 3.12 or older, which has no tag, is pushed under its chart version only. A `+` in TAG is pushed as `_`.")

	return cmd
}

func runCopy(ctx context.Context) error {
	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                &commonCmdData,
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

	if cmdData.From == "" {
		return fmt.Errorf("--from=ADDRESS param required")
	}

	if cmdData.To == "" {
		return fmt.Errorf("--to=ADDRESS param required")
	}

	return logboek.Context(ctx).LogProcess("Copy bundle").DoError(func() error {
		logboek.Context(ctx).LogFDetails("From: %s\n", cmdData.From)
		logboek.Context(ctx).LogFDetails("To: %s\n", cmdData.To)

		return bundles.Copy(ctx, cmdData.From, cmdData.To, bundles.CopyOptions{
			ChartRepoConnectionOptions: common.GetBundlesRegistryConnectionOptions(&commonCmdData),
			RegistryCredentialsPath:    docker.GetDockerConfigCredentialsFile(*commonCmdData.DockerConfig),
			TempDirPath:                werf.GetTmpDir(),
		})
	})
}
