package secret

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "generate-secret-key",
		DisableFlagsInUseLine: true,
		Short:                 "Generate hex encryption key",
		Long: common.GetLongCommandDescription(`Generate hex encryption key.
For further usage, the encryption key should be saved in $WERF_SECRET_KEY or .werf_secret_key file`),
		Example: `  # Export encryption key
  $ export WERF_SECRET_KEY=$(werf helm secret generate-secret-key)

  # Save encryption key in .werf_secret_key file
  $ werf helm secret generate-secret-key > .werf_secret_key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return runGenerateSecretKey()
		},
	})

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func runGenerateSecretKey() error {
	key, err := secrets_manager.GenerateSecretKey()
	if err != nil {
		return err
	}

	fmt.Println(string(key))

	return nil
}
