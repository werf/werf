package main

import (
	"os"

	"github.com/flant/dapp/cmd/dapp/bp"
	"github.com/flant/dapp/cmd/dapp/build"
	"github.com/flant/dapp/cmd/dapp/cleanup"
	"github.com/flant/dapp/cmd/dapp/deploy"
	"github.com/flant/dapp/cmd/dapp/dismiss"
	"github.com/flant/dapp/cmd/dapp/flush"
	"github.com/flant/dapp/cmd/dapp/lint"
	"github.com/flant/dapp/cmd/dapp/push"
	"github.com/flant/dapp/cmd/dapp/render"
	"github.com/flant/dapp/cmd/dapp/reset"
	"github.com/flant/dapp/cmd/dapp/secret"
	"github.com/flant/dapp/cmd/dapp/sync"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:          "dapp",
		SilenceUsage: true,
	}

	cmd.AddCommand(
		build.NewCmd(),
		push.NewCmd(),
		bp.NewCmd(),

		deploy.NewCmd(),
		dismiss.NewCmd(),
		lint.NewCmd(),
		render.NewCmd(),

		reset.NewCmd(),
		flush.NewCmd(),
		sync.NewCmd(),
		cleanup.NewCmd(),

		secretCmd(),
	)

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func secretCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "secret"}
	cmd.AddCommand(
		secret.NewKeyGenCmd(),
		secret.NewGenerateCmd(),
		secret.NewExtractCmd(),
		secret.NewEditCmd(),
		secret.NewRegenerateCmd(),
	)

	return cmd
}
