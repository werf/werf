package download

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"

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

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "download",
		Short:                 "Download published bundle into directory",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag or version mask and unpack it into provided directory (or into directory named as a resulting chart in the current working directory).`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer global_warnings.PrintGlobalWarnings(common.GetContext())

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(runDownload)
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupRepoOptions(&commonCmdData, cmd) // FIXME
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide exact tag version or semver-based pattern, werf will install or upgrade to the latest version of the specified bundle ($WERF_TAG or latest by default)")
	cmd.Flags().StringVarP(&cmdData.Destination, "destination", "d", os.Getenv("WERF_DESTINATION"), "Download bundle into the provided directory ($WERF_DESTINATION or chart-name by default)")

	return cmd
}

func runDownload() error {
	ctx := common.GetContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	repoAddress, err := common.GetStagesStorageAddress(&commonCmdData)
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
