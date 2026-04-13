package merge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/sbom/merge"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)

	var opts merge.Options

	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "merge",
		Short:                 "Merge per-image SBOMs into a product-level SBOM",
		Long:                  common.GetLongCommandDescription(GetDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.DocsLongMD: GetDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := merge.ValidateOptions(opts); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runMerge(ctx, opts)
			})
		},
	})

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{OptionalRepo: false})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to pull SBOM images from the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	setupMergeFlags(cmd, &opts)

	return cmd
}

func setupMergeFlags(cmd *cobra.Command, opts *merge.Options) {
	flags := cmd.Flags()

	flags.StringVarP(&opts.Input, "input", "", "", "Path to JSON mapping file (image name -> sha256:digest)")
	flags.StringVarP(&opts.IsprasFormat, "ispras-format", "", "", `ISPRAS SBOM format: "oss" or "container"`)
	flags.StringVarP(&opts.AppName, "app-name", "", "", "Application/product name for the merged SBOM metadata")
	flags.StringVarP(&opts.AppVersion, "app-version", "", "", "Application/product version for the merged SBOM metadata")
	flags.StringVarP(&opts.Manufacturer, "manufacturer", "", "", "Manufacturer name for the merged SBOM metadata")
	flags.StringVarP(&opts.Output, "output", "o", "", "Output file path (defaults to stdout)")
}

func runMerge(ctx context.Context, opts merge.Options) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)

	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd:                &commonCmdData,
		InitWerf:           true,
		InitDockerRegistry: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	repoAddr, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return err
	}

	registry, err := common.CreateDockerRegistry(ctx, repoAddr, *commonCmdData.InsecureRegistry, *commonCmdData.SkipTlsVerifyRegistry)
	if err != nil {
		return fmt.Errorf("create docker registry: %w", err)
	}

	return merge.Run(ctx, registry, repoAddr, opts)
}
