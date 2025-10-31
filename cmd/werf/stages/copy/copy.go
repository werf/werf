package copy

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	helm_v3 "github.com/werf/3p-helm-for-werf-helm/cmd/helm"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
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

	if cmdData.From != "" {
		return fmt.Errorf("--from=ADDRESS param required")
	}

	if cmdData.To == "" {
		return fmt.Errorf("--to ADDRESS param required")
	}

	fromAddrRaw := cmdData.From
	toAddrRaw := cmdData.To

	logboek.Context(ctx).Debug().LogF("--- START ---\n")
	logboek.Context(ctx).Debug().LogF("FROM_ADDR_RAW: %s\n", fromAddrRaw)
	logboek.Context(ctx).Debug().LogF("TO_ADDR_RAW: %s\n", toAddrRaw)
	logboek.Context(ctx).Debug().LogF("--- END ---\n")

	return nil
}
