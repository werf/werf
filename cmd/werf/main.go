package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flant/werf/cmd/werf/bp"
	"github.com/flant/werf/cmd/werf/build"
	"github.com/flant/werf/cmd/werf/cleanup"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/templates"
	"github.com/flant/werf/cmd/werf/completion"
	"github.com/flant/werf/cmd/werf/deploy"
	"github.com/flant/werf/cmd/werf/dismiss"
	"github.com/flant/werf/cmd/werf/docs"
	"github.com/flant/werf/cmd/werf/flush"
	"github.com/flant/werf/cmd/werf/gc"
	"github.com/flant/werf/cmd/werf/lint"
	"github.com/flant/werf/cmd/werf/push"
	"github.com/flant/werf/cmd/werf/render"
	"github.com/flant/werf/cmd/werf/reset"
	"github.com/flant/werf/cmd/werf/sync"
	"github.com/flant/werf/cmd/werf/tag"
	"github.com/flant/werf/cmd/werf/version"
	"github.com/flant/werf/pkg/logger"
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
		logger.LogError(fmt.Errorf("process exterminator initialization error: %s", err))
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "werf",
		Short: "Werf helps to implement and support Continuous Integration and Continuous Delivery",
		Long: common.GetLongCommandDescription(`Werf helps to implement and support Continuous Integration and Continuous Delivery.

Find more information at https://flant.github.io/werf`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	groups := templates.CommandGroups{
		{
			Message: "Build Commands:",
			Commands: []*cobra.Command{
				build.NewCmd(),
				push.NewCmd(),
				bp.NewCmd(),
				tag.NewCmd(),
			},
		},
		{
			Message: "Deploy Commands:",
			Commands: []*cobra.Command{
				deploy.NewCmd(),
				dismiss.NewCmd(),
				lint.NewCmd(),
				render.NewCmd(),
				secretCmd(),
			},
		},
		{
			Message: "Project Cleanup Commands:",
			Commands: []*cobra.Command{
				flush.NewCmd(),
				sync.NewCmd(),
				cleanup.NewCmd(),
			},
		},
		{
			Message: "Cleanup Commands:",
			Commands: []*cobra.Command{
				reset.NewCmd(),
				gc.NewCmd(),
			},
		},
	}
	groups.Add(rootCmd)

	templates.ActsAsRootCommand(rootCmd, groups...)

	rootCmd.AddCommand(
		slugCmd(),
		completion.NewCmd(rootCmd),
		version.NewCmd(),
		docs.NewCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		logger.LogError(err)
		os.Exit(1)
	}
}

func secretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Commands to work with secrets",
	}
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

		logger.LogError(errors.New("interrupted"))
		os.Exit(17)
	}()
}
