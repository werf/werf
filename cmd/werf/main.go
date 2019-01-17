package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flant/werf/cmd/werf/bp"
	"github.com/flant/werf/cmd/werf/build"
	"github.com/flant/werf/cmd/werf/cleanup"
	"github.com/flant/werf/cmd/werf/completion"
	"github.com/flant/werf/cmd/werf/deploy"
	"github.com/flant/werf/cmd/werf/dismiss"
	"github.com/flant/werf/cmd/werf/flush"
	"github.com/flant/werf/cmd/werf/gc"
	"github.com/flant/werf/cmd/werf/lint"
	"github.com/flant/werf/cmd/werf/push"
	"github.com/flant/werf/cmd/werf/render"
	"github.com/flant/werf/cmd/werf/reset"
	"github.com/flant/werf/cmd/werf/sync"
	"github.com/flant/werf/cmd/werf/tag"
	"github.com/flant/werf/cmd/werf/version"
	"github.com/flant/werf/pkg/process_exterminator"

	secret_edit "github.com/flant/werf/cmd/werf/secret/edit"
	secret_extract "github.com/flant/werf/cmd/werf/secret/extract"
	secret_generate "github.com/flant/werf/cmd/werf/secret/generate"
	secret_key_generate "github.com/flant/werf/cmd/werf/secret/key_generate"
	secret_regenerate "github.com/flant/werf/cmd/werf/secret/regenerate"

	slug_namespace "github.com/flant/werf/cmd/werf/slug/namespace"
	slug_release "github.com/flant/werf/cmd/werf/slug/release"
	slug_tag "github.com/flant/werf/cmd/werf/slug/tag"

	"github.com/spf13/cobra"
)

func main() {
	trapTerminationSignals()

	if err := process_exterminator.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Process exterminator initialization error: %s\n", err)
		os.Exit(1)
	}

	cmd := &cobra.Command{
		Use:          "werf",
		SilenceUsage: true,
	}

	cmd.AddCommand(
		build.NewCmd(),
		push.NewCmd(),
		bp.NewCmd(),
		tag.NewCmd(),

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

		completion.NewCmd(cmd),
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
