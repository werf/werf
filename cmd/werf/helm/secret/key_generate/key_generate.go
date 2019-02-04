package secret

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy/secret"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "keygen",
		DisableFlagsInUseLine: true,
		Short: "Generate hex encryption key that can be used as WERF_SECRET_KEY",
		Long: common.GetLongCommandDescription(`Generate hex key that can be used as WERF_SECRET_KEY.

16-bytes key will be generated (AES-128).`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSecretKeyGenerate()
		},
	}

	return cmd
}

func runSecretKeyGenerate() error {
	key, err := secret.GenerateSecretKey()
	if err != nil {
		return err
	}

	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf("WERF_SECRET_KEY=%s\n", string(key))
	} else {
		fmt.Println(string(key))
	}

	return nil
}
