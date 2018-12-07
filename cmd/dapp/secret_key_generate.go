package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/flant/dapp/pkg/deploy/secret"
)

func newSecretKeyGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate encryption key",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runSecretKeyGenerate()
			if err != nil {
				return fmt.Errorf("secret keygen failed: %s", err)
			}
			return nil
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
		fmt.Printf("DAPP_SECRET_KEY=%s\n", string(key))
	} else {
		fmt.Println(string(key))
	}

	return nil
}
