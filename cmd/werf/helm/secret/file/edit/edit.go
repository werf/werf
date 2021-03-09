package secret

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/deploy/secrets_manager"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	secret_common "github.com/werf/werf/cmd/werf/helm/secret/common"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/werf"
)

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "edit FILE_PATH",
		DisableFlagsInUseLine: true,
		Short:                 "Edit or create new secret file",
		Long: common.GetLongCommandDescription(`Edit or create new secret file.
Encryption key should be in $WERF_SECRET_KEY or .werf_secret_key file`),
		Example: `  # Create/edit existing secret file
  $ werf helm secret file edit .helm/secret/privacy`,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
				return err
			}

			return runSecretEdit(common.BackgroundContext(), args[0])
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func runSecretEdit(ctx context.Context, filePath string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	workingDir := common.GetWorkingDir(&commonCmdData)

	return secret_common.SecretEdit(ctx, secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{}), workingDir, filePath, false)
}
