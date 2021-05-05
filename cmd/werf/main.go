package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/build"
	"github.com/werf/werf/cmd/werf/ci_env"
	"github.com/werf/werf/cmd/werf/cleanup"
	"github.com/werf/werf/cmd/werf/compose"
	"github.com/werf/werf/cmd/werf/converge"
	"github.com/werf/werf/cmd/werf/dismiss"
	"github.com/werf/werf/cmd/werf/helm"
	"github.com/werf/werf/cmd/werf/purge"
	"github.com/werf/werf/cmd/werf/run"
	"github.com/werf/werf/cmd/werf/slugify"
	"github.com/werf/werf/cmd/werf/synchronization"

	managed_images_add "github.com/werf/werf/cmd/werf/managed_images/add"
	managed_images_ls "github.com/werf/werf/cmd/werf/managed_images/ls"
	managed_images_rm "github.com/werf/werf/cmd/werf/managed_images/rm"

	host_cleanup "github.com/werf/werf/cmd/werf/host/cleanup"
	host_purge "github.com/werf/werf/cmd/werf/host/purge"

	bundle_apply "github.com/werf/werf/cmd/werf/bundle/apply"
	bundle_download "github.com/werf/werf/cmd/werf/bundle/download"
	bundle_export "github.com/werf/werf/cmd/werf/bundle/export"
	bundle_publish "github.com/werf/werf/cmd/werf/bundle/publish"

	config_list "github.com/werf/werf/cmd/werf/config/list"
	config_render "github.com/werf/werf/cmd/werf/config/render"
	"github.com/werf/werf/cmd/werf/render"

	"github.com/werf/werf/cmd/werf/completion"
	"github.com/werf/werf/cmd/werf/docs"
	"github.com/werf/werf/cmd/werf/version"

	stage_image "github.com/werf/werf/cmd/werf/stage/image"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/common/templates"
	"github.com/werf/werf/pkg/process_exterminator"
)

func main() {
	common.EnableTerminationSignalsTrap()
	log.SetOutput(logboek.OutStream())
	logrus.StandardLogger().SetOutput(logboek.OutStream())

	if err := process_exterminator.Init(); err != nil {
		common.TerminateWithError(fmt.Sprintf("process exterminator initialization failed: %s", err), 1)
	}

	rootCmd := constructRootCmd()

	if err := rootCmd.Execute(); err != nil {
		common.TerminateWithError(err.Error(), 1)
	}
}

func constructRootCmd() *cobra.Command {
	if filepath.Base(os.Args[0]) == "helm" || os.Getenv("WERF_HELM3_MODE") == "1" {
		return helm.NewCmd()
	}

	rootCmd := &cobra.Command{
		Use:   "werf",
		Short: "werf helps to implement and support Continuous Integration and Continuous Delivery",
		Long: common.GetLongCommandDescription(`werf helps to implement and support Continuous Integration and Continuous Delivery.

Find more information at https://werf.io`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	groups := &templates.CommandGroups{}
	*groups = append(*groups, templates.CommandGroups{
		{
			Message: "Delivery commands",
			Commands: []*cobra.Command{
				converge.NewCmd(),
				dismiss.NewCmd(),
				bundleCmd(),
			},
		},
		{
			Message: "Cleaning commands",
			Commands: []*cobra.Command{
				cleanup.NewCmd(),
				purge.NewCmd(),
			},
		},
		{
			Message: "Helper commands",
			Commands: []*cobra.Command{
				ci_env.NewCmd(),
				build.NewCmd(),
				run.NewCmd(),
				dockerComposeCmd(),
				slugify.NewCmd(),
				render.NewCmd(),
			},
		},
		{
			Message: "Low-level management commands",
			Commands: []*cobra.Command{
				configCmd(),
				managedImagesCmd(),
				hostCmd(),
				helm.NewCmd(),
			},
		},
		{
			Message: "Other commands",
			Commands: []*cobra.Command{
				synchronization.NewCmd(),
				completion.NewCmd(rootCmd),
				version.NewCmd(),
				docs.NewCmd(groups),
				stageCmd(),
			},
		},
	}...)
	groups.Add(rootCmd)

	templates.ActsAsRootCommand(rootCmd, *groups...)

	return rootCmd
}

func dockerComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Work with docker-compose",
	}
	cmd.AddCommand(
		compose.NewConfigCmd(),
		compose.NewRunCmd(),
		compose.NewUpCmd(),
		compose.NewDownCmd(),
	)

	return cmd

}

func bundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Work with werf bundles: publish bundles into container registry and deploy bundles into Kubernetes cluster",
	}
	cmd.AddCommand(
		bundle_publish.NewCmd(),
		bundle_apply.NewCmd(),
		bundle_export.NewCmd(),
		bundle_download.NewCmd(),
	)

	return cmd
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Work with werf.yaml",
	}
	cmd.AddCommand(
		config_render.NewCmd(),
		config_list.NewCmd(),
	)

	return cmd
}

func managedImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "managed-images",
		Short: "Work with managed images which will be preserved during cleanup procedure",
	}
	cmd.AddCommand(
		managed_images_add.NewCmd(),
		managed_images_ls.NewCmd(),
		managed_images_rm.NewCmd(),
	)

	return cmd
}

func stageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "stage",
		Hidden: true,
	}
	cmd.AddCommand(
		stage_image.NewCmd(),
	)

	return cmd
}

func hostCmd() *cobra.Command {
	hostCmd := &cobra.Command{
		Use:   "host",
		Short: "Work with werf cache and data of all projects on the host machine",
	}

	hostCmd.AddCommand(
		host_cleanup.NewCmd(),
		host_purge.NewCmd(),
	)

	return hostCmd
}
