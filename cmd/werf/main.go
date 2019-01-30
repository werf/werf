package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/cleanup"

	secret_edit "github.com/flant/werf/cmd/werf/secret/edit"
	secret_extract "github.com/flant/werf/cmd/werf/secret/extract"
	secret_generate "github.com/flant/werf/cmd/werf/secret/generate"
	secret_key_generate "github.com/flant/werf/cmd/werf/secret/key_generate"
	secret_regenerate "github.com/flant/werf/cmd/werf/secret/regenerate"

	slug_namespace "github.com/flant/werf/cmd/werf/slug/namespace"
	slug_release "github.com/flant/werf/cmd/werf/slug/release"
	slug_tag "github.com/flant/werf/cmd/werf/slug/tag"

	images_cleanup "github.com/flant/werf/cmd/werf/images/cleanup"
	images_publish "github.com/flant/werf/cmd/werf/images/publish"

	stages_build "github.com/flant/werf/cmd/werf/stages/build"
	stages_cleanup "github.com/flant/werf/cmd/werf/stages/cleanup"

	"github.com/flant/werf/cmd/werf/completion"
	"github.com/flant/werf/cmd/werf/docs"
	"github.com/flant/werf/cmd/werf/version"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/templates"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/process_exterminator"
)

func main() {
	trapTerminationSignals()

	logger.Init()

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
			Message: "Main Commands:",
			Commands: []*cobra.Command{
				cleanup.NewCmd(),
			},
		},
		{
			Message: "Lowlevel Management Commands:",
			Commands: []*cobra.Command{
				stagesCmd(),
				imagesCmd(),
			},
		},
	}
	groups.Add(rootCmd)

	templates.ActsAsRootCommand(rootCmd, groups...)

	rootCmd.AddCommand(
		completion.NewCmd(rootCmd),
		version.NewCmd(),
		docs.NewCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		logger.LogError(err)
		os.Exit(1)
	}
}

func imagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Commands to work with images",
	}
	cmd.AddCommand(
		images_publish.NewCmd(),
		images_cleanup.NewCmd(),
	)

	return cmd
}

func stagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stages",
		Short: "Commands to work with stages, which are cache for images",
	}
	cmd.AddCommand(
		stages_build.NewCmd(),
		stages_cleanup.NewCmd(),
	)

	return cmd
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
