package copy

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
	Repo  *common.RepoData
	Tag   string
	To    *common.RepoData
	ToTag string
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "copy",
		Short:                 "Copy published bundle into another location",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag and copy it either into a different tag within the same container registry or into another container registry.`),
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

			return common.LogRunningTime(runCopy)
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repos")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	cmdData.Repo = common.NewRepoData("repo", common.RepoDataOptions{OnlyAddress: true})
	cmdData.Repo.SetupCmd(cmd)

	cmdData.To = common.NewRepoData("to", common.RepoDataOptions{OnlyAddress: true})
	cmdData.To.SetupCmd(cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)
	common.SetupPlatform(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide from tag version of the bundle to copy ($WERF_TAG or latest by default)")
	cmd.Flags().StringVarP(&cmdData.ToTag, "to-tag", "", "", "Provide to tag version of the bundle to copy ($WERF_TO_TAG or same as --tag by default)")

	return cmd
}

func runCopy() error {
	ctx := common.GetContext()

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

	if *cmdData.Repo.Address == "" {
		return fmt.Errorf("--repo=ADDRESS param required")
	}
	if *cmdData.To.Address == "" {
		return fmt.Errorf("--to=ADDRESS param required")
	}

	fromRegistry, err := cmdData.Repo.CreateDockerRegistry(*commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
	if err != nil {
		return fmt.Errorf("error creating container registry accessor for repo %s: %w", *cmdData.Repo.Address, err)
	}

	fromTag := cmdData.Tag
	toTag := cmdData.ToTag
	if toTag == "" {
		toTag = fromTag
	}

	fromRef := fmt.Sprintf("%s:%s", *cmdData.Repo.Address, fromTag)
	toRef := fmt.Sprintf("%s:%s", *cmdData.To.Address, toTag)

	return bundles.Copy(ctx, fromRef, toRef, bundlesRegistryClient, fromRegistry)
}
