package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/build"
	"github.com/flant/werf/cmd/werf/build_and_publish"
	"github.com/flant/werf/cmd/werf/cleanup"
	"github.com/flant/werf/cmd/werf/deploy"
	"github.com/flant/werf/cmd/werf/dismiss"
	"github.com/flant/werf/cmd/werf/publish"
	"github.com/flant/werf/cmd/werf/purge"
	"github.com/flant/werf/cmd/werf/run"

	helm_secret_edit "github.com/flant/werf/cmd/werf/helm/secret/edit"
	helm_secret_extract "github.com/flant/werf/cmd/werf/helm/secret/extract"
	helm_secret_generate "github.com/flant/werf/cmd/werf/helm/secret/generate"
	helm_secret_key_generate "github.com/flant/werf/cmd/werf/helm/secret/key_generate"
	helm_secret_regenerate "github.com/flant/werf/cmd/werf/helm/secret/regenerate"

	"github.com/flant/werf/cmd/werf/ci_env"
	"github.com/flant/werf/cmd/werf/slugify"

	images_cleanup "github.com/flant/werf/cmd/werf/images/cleanup"
	images_publish "github.com/flant/werf/cmd/werf/images/publish"
	images_purge "github.com/flant/werf/cmd/werf/images/purge"

	stages_build "github.com/flant/werf/cmd/werf/stages/build"
	stages_cleanup "github.com/flant/werf/cmd/werf/stages/cleanup"
	stages_purge "github.com/flant/werf/cmd/werf/stages/purge"

	host_cleanup "github.com/flant/werf/cmd/werf/host/cleanup"
	host_purge "github.com/flant/werf/cmd/werf/host/purge"

	helm_deploy_chart "github.com/flant/werf/cmd/werf/helm/deploy_chart"
	helm_generate_chart "github.com/flant/werf/cmd/werf/helm/generate_chart"
	helm_get_service_values "github.com/flant/werf/cmd/werf/helm/get_service_values"
	helm_lint "github.com/flant/werf/cmd/werf/helm/lint"
	helm_render "github.com/flant/werf/cmd/werf/helm/render"

	meta_get_helm_release "github.com/flant/werf/cmd/werf/meta/get_helm_release"
	meta_get_namespace "github.com/flant/werf/cmd/werf/meta/get_namespace"

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

	logger.Init(logger.Options{})

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
				build.NewCmd(),
				publish.NewCmd(),
				build_and_publish.NewCmd(),
				run.NewCmd(),
				deploy.NewCmd(),
				dismiss.NewCmd(),
				cleanup.NewCmd(),
				purge.NewCmd(),
			},
		},
		{
			Message: "Toolbox:",
			Commands: []*cobra.Command{
				slugify.NewCmd(),
				ci_env.NewCmd(),
				metaCmd(),
			},
		},
		{
			Message: "Lowlevel Management Commands:",
			Commands: []*cobra.Command{
				stagesCmd(),
				imagesCmd(),
				helmCmd(),
				hostCmd(),
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
		logger.LogError(fmt.Errorf("Error: %s", err))
		os.Exit(1)
	}
}

func imagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Work with images",
	}
	cmd.AddCommand(
		images_publish.NewCmd(),
		images_cleanup.NewCmd(),
		images_purge.NewCmd(),
	)

	return cmd
}

func stagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stages",
		Short: "Work with stages, which are cache for images",
	}
	cmd.AddCommand(
		stages_build.NewCmd(),
		stages_cleanup.NewCmd(),
		stages_purge.NewCmd(),
	)

	return cmd
}

func helmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm",
		Short: "Manage application deployment with helm",
	}
	cmd.AddCommand(
		helm_get_service_values.NewCmd(),
		helm_generate_chart.NewCmd(),
		helm_deploy_chart.NewCmd(),
		helm_lint.NewCmd(),
		helm_render.NewCmd(),
		secretCmd(),
	)

	return cmd
}

func hostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "host",
	}
	cmd.AddCommand(
		host_cleanup.NewCmd(),
		host_purge.NewCmd(),
	)

	return cmd
}

func secretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Work with secrets",
	}
	cmd.AddCommand(
		helm_secret_key_generate.NewCmd(),
		helm_secret_generate.NewCmd(),
		helm_secret_extract.NewCmd(),
		helm_secret_edit.NewCmd(),
		helm_secret_regenerate.NewCmd(),
	)

	return cmd
}

func metaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta",
		Short: "Work with werf project meta configuration",
	}
	cmd.AddCommand(
		meta_get_helm_release.NewCmd(),
		meta_get_namespace.NewCmd(),
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
