package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flant/dapp/cmd/dapp/bp"
	"github.com/flant/dapp/cmd/dapp/build"
	"github.com/flant/dapp/cmd/dapp/cleanup"
	"github.com/flant/dapp/cmd/dapp/deploy"
	"github.com/flant/dapp/cmd/dapp/dismiss"
	"github.com/flant/dapp/cmd/dapp/flush"
	"github.com/flant/dapp/cmd/dapp/gc"
	"github.com/flant/dapp/cmd/dapp/lint"
	"github.com/flant/dapp/cmd/dapp/push"
	"github.com/flant/dapp/cmd/dapp/render"
	"github.com/flant/dapp/cmd/dapp/reset"
	"github.com/flant/dapp/cmd/dapp/sync"
	"github.com/flant/dapp/cmd/dapp/version"
	"github.com/flant/dapp/pkg/process_exterminator"

	secret_edit "github.com/flant/dapp/cmd/dapp/secret/edit"
	secret_extract "github.com/flant/dapp/cmd/dapp/secret/extract"
	secret_generate "github.com/flant/dapp/cmd/dapp/secret/generate"
	secret_key_generate "github.com/flant/dapp/cmd/dapp/secret/key_generate"
	secret_regenerate "github.com/flant/dapp/cmd/dapp/secret/regenerate"

	slug_namespace "github.com/flant/dapp/cmd/dapp/slug/namespace"
	slug_release "github.com/flant/dapp/cmd/dapp/slug/release"
	slug_tag "github.com/flant/dapp/cmd/dapp/slug/tag"

	"github.com/spf13/cobra"
)

func main() {
	trapTerminationSignals()

	if err := process_exterminator.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Process exterminator initialization error: %s\n", err)
		os.Exit(1)
	}

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
		gc.NewCmd(),

		secretCmd(),
		slugCmd(),

		version.NewCmd(),
	)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func secretCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "secret"}
	cmd.AddCommand(
		secret_key_generate.NewCmd(),
		secret_generate.NewCmd(),
		secret_extract.NewCmd(),
		secret_edit.NewCmd(),
		secret_regenerate.NewCmd(),
	)

	return cmd
}

func slugCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "slug"}
	cmd.AddCommand(
		slug_tag.NewCmd(),
		slug_namespace.NewCmd(),
		slug_release.NewCmd(),
	)

	return cmd
}

func trapTerminationSignals() {
	c := make(chan os.Signal, 1)
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT}
	signal.Notify(c, signals...)
	go func() {
		<-c

		fmt.Fprintf(os.Stderr, "Interrupted\n")

		os.Exit(17)
	}()
}
