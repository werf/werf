package download

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/bundles"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag         string
	Destination string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "download",
		Short:                 "Download published bundle into directory",
		Hidden:                true, // Deprecated command
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag or version mask and unpack it into provided directory (or into directory named as a resulting chart in the current working directory).`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			logboek.Context(ctx).Warn().LogF("WARNING: werf bundle download is DEPRECATED!\n")
			logboek.Context(ctx).Warn().LogF("WARNING: To download published bundle from the registry and unpack bundle helm chart into directory use following commands:\n")
			logboek.Context(ctx).Warn().LogF("WARNING: \n")
			logboek.Context(ctx).Warn().LogF("WARNING: 1. Publish bundle into some registry:\n")
			logboek.Context(ctx).Warn().LogF("WARNING:     werf bundle publish --repo REPO --tag TAG\n")
			logboek.Context(ctx).Warn().LogF("WARNING: \n")
			logboek.Context(ctx).Warn().LogF("WARNING: 2. Copy published bundle from the registry to bundle archive:\n")
			logboek.Context(ctx).Warn().LogF("WARNING:     werf bundle copy --from REPO:TAG --to archive:PATH_TO_ARCHIVE.tar.gz\n")
			logboek.Context(ctx).Warn().LogF("WARNING: \n")
			logboek.Context(ctx).Warn().LogF("WARNING: 3. Unpack bundle archive:\n")
			logboek.Context(ctx).Warn().LogF("WARNING:     tar xf PATH_TO_ARCHIVE.tar.gz\n")
			logboek.Context(ctx).Warn().LogF("WARNING: \n")
			logboek.Context(ctx).Warn().LogF("WARNING: 4. Unpack chart archive inside unpacked bundle archive directory:\n")
			logboek.Context(ctx).Warn().LogF("WARNING:     tar xf chart.tar.gz\n")

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runDownload(ctx) })
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "")

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide exact tag version or semver-based pattern, werf will install or upgrade to the latest version of the specified bundle ($WERF_TAG or latest by default)")
	cmd.Flags().StringVarP(&cmdData.Destination, "destination", "d", os.Getenv("WERF_DESTINATION"), "Download bundle into the provided directory ($WERF_DESTINATION or chart-name by default)")

	return cmd
}

func runDownload(ctx context.Context) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	repoAddress, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return err
	}

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	return bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), cmdData.Destination, bundlesRegistryClient)
}
