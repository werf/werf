package secret

import (
	"fmt"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
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
			return runGenerateSecretKey()
		},
	}

	return cmd
}

func runGenerateSecretKey() error {
	key, err := secret.GenerateSecretKey()
	if err != nil {
		return err
	}

	fmt.Println(string(key))

	return nil
}
