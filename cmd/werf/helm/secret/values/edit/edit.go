package secret

import (
	"cmp"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/nelm/pkg/action"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/docs/replacers/helm"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "edit FILE_PATH",
		DisableFlagsInUseLine: true,
		Short:                 "Edit or create new secret values file",
		Long:                  common.GetLongCommandDescription(helm.GetHelmSecretValuesEditDocs().Long),
		Example: `  # Create/edit existing secret values file
  $ werf helm secret values edit .helm/secret-values.yaml`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
			common.DocsLongMD: helm.GetHelmSecretValuesEditDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
				return err
			}

			return runSecretEdit(ctx, args[0])
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func runSecretEdit(ctx context.Context, filepPath string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	workingDir := common.GetWorkingDir(&commonCmdData)

	ctx = log.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultSecretValuesFileEditLogLevel), log.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})

	if err := action.SecretValuesFileEdit(ctx, filepPath, action.SecretValuesFileEditOptions{
		TempDirPath:   werf.GetTmpDir(),
		SecretWorkDir: workingDir,
	}); err != nil {
		return fmt.Errorf("secret values file edit: %w", err)
	}

	return nil
}
