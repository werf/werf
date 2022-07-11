package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/build"
	bundle_apply "github.com/werf/werf/cmd/werf/bundle/apply"
	bundle_copy "github.com/werf/werf/cmd/werf/bundle/copy"
	bundle_download "github.com/werf/werf/cmd/werf/bundle/download"
	bundle_export "github.com/werf/werf/cmd/werf/bundle/export"
	bundle_publish "github.com/werf/werf/cmd/werf/bundle/publish"
	bundle_render "github.com/werf/werf/cmd/werf/bundle/render"
	"github.com/werf/werf/cmd/werf/ci_env"
	"github.com/werf/werf/cmd/werf/cleanup"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/common/templates"
	"github.com/werf/werf/cmd/werf/completion"
	"github.com/werf/werf/cmd/werf/compose"
	config_graph "github.com/werf/werf/cmd/werf/config/graph"
	config_list "github.com/werf/werf/cmd/werf/config/list"
	config_render "github.com/werf/werf/cmd/werf/config/render"
	"github.com/werf/werf/cmd/werf/converge"
	cr_login "github.com/werf/werf/cmd/werf/cr/login"
	cr_logout "github.com/werf/werf/cmd/werf/cr/logout"
	"github.com/werf/werf/cmd/werf/dismiss"
	"github.com/werf/werf/cmd/werf/docs"
	"github.com/werf/werf/cmd/werf/export"
	"github.com/werf/werf/cmd/werf/helm"
	host_cleanup "github.com/werf/werf/cmd/werf/host/cleanup"
	host_purge "github.com/werf/werf/cmd/werf/host/purge"
	"github.com/werf/werf/cmd/werf/kube_run"
	"github.com/werf/werf/cmd/werf/kubectl"
	managed_images_add "github.com/werf/werf/cmd/werf/managed_images/add"
	managed_images_ls "github.com/werf/werf/cmd/werf/managed_images/ls"
	managed_images_rm "github.com/werf/werf/cmd/werf/managed_images/rm"
	"github.com/werf/werf/cmd/werf/purge"
	"github.com/werf/werf/cmd/werf/render"
	"github.com/werf/werf/cmd/werf/run"
	"github.com/werf/werf/cmd/werf/slugify"
	stage_image "github.com/werf/werf/cmd/werf/stage/image"
	"github.com/werf/werf/cmd/werf/synchronization"
	"github.com/werf/werf/cmd/werf/version"
	"github.com/werf/werf/pkg/process_exterminator"
	"github.com/werf/werf/pkg/telemetry"
)

func main() {
	ctx := common.GetContextWithLogger()

	common.InitTelemetry(ctx)

	shouldTerminate, err := common.ContainerBackendProcessStartupHook()
	if err != nil {
		common.ShutdownTelemetry(ctx, 1)
		common.TerminateWithError(err.Error(), 1)
	}
	if shouldTerminate {
		return
	}

	common.EnableTerminationSignalsTrap()
	log.SetOutput(logboek.OutStream())
	logrus.StandardLogger().SetOutput(logboek.OutStream())

	if err := process_exterminator.Init(); err != nil {
		common.ShutdownTelemetry(ctx, 1)
		common.TerminateWithError(fmt.Sprintf("process exterminator initialization failed: %s", err), 1)
	}

	rootCmd := constructRootCmd(ctx)

	setupTelemetryInit(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		if helm_v3.IsPluginError(err) {
			common.ShutdownTelemetry(ctx, helm_v3.PluginErrorCode(err))
			common.TerminateWithError(err.Error(), helm_v3.PluginErrorCode(err))
		} else {
			common.ShutdownTelemetry(ctx, 1)
			common.TerminateWithError(err.Error(), 1)
		}
	}

	common.ShutdownTelemetry(ctx, 0)
}

func constructRootCmd(ctx context.Context) *cobra.Command {
	if filepath.Base(os.Args[0]) == "helm" || helm.IsHelm3Mode() {
		return helm.NewCmd(ctx)
	}

	rootCmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "werf",
		Short: "werf helps to implement and support Continuous Integration and Continuous Delivery",
		Long: common.GetLongCommandDescription(`werf helps to implement and support Continuous Integration and Continuous Delivery.

Find more information at https://werf.io`),
		SilenceUsage:  true,
		SilenceErrors: true,
	})

	groups := &templates.CommandGroups{}
	*groups = append(*groups, templates.CommandGroups{
		{
			Message: "Delivery commands",
			Commands: []*cobra.Command{
				converge.NewCmd(ctx),
				dismiss.NewCmd(ctx),
				bundleCmd(ctx),
			},
		},
		{
			Message: "Cleaning commands",
			Commands: []*cobra.Command{
				cleanup.NewCmd(ctx),
				purge.NewCmd(ctx),
			},
		},
		{
			Message: "Helper commands",
			Commands: []*cobra.Command{
				ci_env.NewCmd(ctx),
				build.NewCmd(ctx),
				export.NewExportCmd(ctx),
				run.NewCmd(ctx),
				kube_run.NewCmd(ctx),
				dockerComposeCmd(ctx),
				slugify.NewCmd(ctx),
				render.NewCmd(ctx),
			},
		},
		{
			Message: "Low-level management commands",
			Commands: []*cobra.Command{
				configCmd(ctx),
				managedImagesCmd(ctx),
				hostCmd(ctx),
				helm.NewCmd(ctx),
				crCmd(ctx),
				kubectl.NewCmd(ctx),
			},
		},
		{
			Message: "Other commands",
			Commands: []*cobra.Command{
				synchronization.NewCmd(ctx),
				completion.NewCmd(ctx, rootCmd),
				version.NewCmd(ctx),
				docs.NewCmd(ctx, groups),
				stageCmd(ctx),
			},
		},
	}...)
	groups.Add(rootCmd)

	templates.ActsAsRootCommand(rootCmd, *groups...)

	return rootCmd
}

func dockerComposeCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "compose",
		Short: "Work with docker-compose",
	})
	cmd.AddCommand(
		compose.NewConfigCmd(ctx),
		compose.NewRunCmd(ctx),
		compose.NewUpCmd(ctx),
		compose.NewDownCmd(ctx),
	)

	return cmd
}

func crCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "cr",
		Short: "Work with container registry: authenticate, list and remove images, etc.",
	})
	cmd.AddCommand(
		cr_login.NewCmd(ctx),
		cr_logout.NewCmd(ctx),
	)

	return cmd
}

func bundleCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "bundle",
		Short: "Work with werf bundles: publish bundles into container registry and deploy bundles into Kubernetes cluster",
	})
	cmd.AddCommand(
		bundle_publish.NewCmd(ctx),
		bundle_apply.NewCmd(ctx),
		bundle_export.NewCmd(ctx),
		bundle_download.NewCmd(ctx),
		bundle_render.NewCmd(ctx),
		bundle_copy.NewCmd(ctx),
	)

	return cmd
}

func configCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "config",
		Short: "Work with werf.yaml",
	})
	cmd.AddCommand(
		config_render.NewCmd(ctx),
		config_list.NewCmd(ctx),
		config_graph.NewCmd(ctx),
	)

	return cmd
}

func managedImagesCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "managed-images",
		Short: "Work with managed images which will be preserved during cleanup procedure",
	})
	cmd.AddCommand(
		managed_images_add.NewCmd(ctx),
		managed_images_ls.NewCmd(ctx),
		managed_images_rm.NewCmd(ctx),
	)

	return cmd
}

func stageCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:    "stage",
		Hidden: true,
	})
	cmd.AddCommand(
		stage_image.NewCmd(ctx),
	)

	return cmd
}

func hostCmd(ctx context.Context) *cobra.Command {
	hostCmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "host",
		Short: "Work with werf cache and data of all projects on the host machine",
	})

	hostCmd.AddCommand(
		host_cleanup.NewCmd(ctx),
		host_purge.NewCmd(ctx),
	)

	return hostCmd
}

func setupTelemetryInit(rootCmd *cobra.Command) {
	commandsQueue := []*cobra.Command{rootCmd}

	for len(commandsQueue) > 0 {
		cmd := commandsQueue[0]
		commandsQueue = commandsQueue[1:]

		commandsQueue = append(commandsQueue, cmd.Commands()...)

		if cmd.Runnable() {
			oldRunE := cmd.RunE
			cmd.RunE = nil

			oldRun := cmd.Run
			cmd.Run = nil

			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				if err := common.TelemetryPreRun(cmd, args); err != nil {
					if telemetry.IsLogsEnabled() {
						fmt.Fprintf(os.Stderr, "Telemetry error: %s\n", err)
					}
				}

				if oldRunE != nil {
					return oldRunE(cmd, args)
				} else if oldRun != nil {
					oldRun(cmd, args)
					return nil
				} else {
					panic(fmt.Sprintf("unexpected command %q, please report bug to the https://github.com/werf/werf", cmd.Name()))
				}
			}
		}
	}
}
