package root

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/cmd/werf/build"
	bundle_apply "github.com/werf/werf/v2/cmd/werf/bundle/apply"
	bundle_copy "github.com/werf/werf/v2/cmd/werf/bundle/copy"
	bundle_publish "github.com/werf/werf/v2/cmd/werf/bundle/publish"
	bundle_render "github.com/werf/werf/v2/cmd/werf/bundle/render"
	"github.com/werf/werf/v2/cmd/werf/ci_env"
	"github.com/werf/werf/v2/cmd/werf/cleanup"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/cmd/werf/common/templates"
	"github.com/werf/werf/v2/cmd/werf/completion"
	"github.com/werf/werf/v2/cmd/werf/compose"
	config_graph "github.com/werf/werf/v2/cmd/werf/config/graph"
	config_list "github.com/werf/werf/v2/cmd/werf/config/list"
	config_render "github.com/werf/werf/v2/cmd/werf/config/render"
	"github.com/werf/werf/v2/cmd/werf/converge"
	cr_login "github.com/werf/werf/v2/cmd/werf/cr/login"
	cr_logout "github.com/werf/werf/v2/cmd/werf/cr/logout"
	"github.com/werf/werf/v2/cmd/werf/dismiss"
	"github.com/werf/werf/v2/cmd/werf/docs"
	kubectl2 "github.com/werf/werf/v2/cmd/werf/docs/replacers/kubectl"
	"github.com/werf/werf/v2/cmd/werf/export"
	"github.com/werf/werf/v2/cmd/werf/helm"
	host_cleanup "github.com/werf/werf/v2/cmd/werf/host/cleanup"
	host_purge "github.com/werf/werf/v2/cmd/werf/host/purge"
	includes_getfile "github.com/werf/werf/v2/cmd/werf/includes/get-file"
	includes_lsfiles "github.com/werf/werf/v2/cmd/werf/includes/ls-files"
	includes_update "github.com/werf/werf/v2/cmd/werf/includes/update"
	"github.com/werf/werf/v2/cmd/werf/kube_run"
	"github.com/werf/werf/v2/cmd/werf/kubectl"
	managed_images_add "github.com/werf/werf/v2/cmd/werf/managed_images/add"
	managed_images_ls "github.com/werf/werf/v2/cmd/werf/managed_images/ls"
	managed_images_rm "github.com/werf/werf/v2/cmd/werf/managed_images/rm"
	"github.com/werf/werf/v2/cmd/werf/plan"
	"github.com/werf/werf/v2/cmd/werf/purge"
	"github.com/werf/werf/v2/cmd/werf/render"
	"github.com/werf/werf/v2/cmd/werf/run"
	"github.com/werf/werf/v2/cmd/werf/slugify"
	stage_image "github.com/werf/werf/v2/cmd/werf/stage/image"
	"github.com/werf/werf/v2/cmd/werf/synchronization"
	"github.com/werf/werf/v2/cmd/werf/version"
	"github.com/werf/werf/v2/pkg/telemetry"
)

func ConstructRootCmd(ctx context.Context) (*cobra.Command, error) {
	helmCmd, err := helm.NewCmd(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to init helm commands: %w", err)
	}

	if filepath.Base(os.Args[0]) == "helm" || helm.IsHelm3Mode() {
		return helmCmd, nil
	}

	rootCmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:           "werf",
		Short:         "werf helps to implement and support Continuous Integration and Continuous Delivery",
		Long:          common.GetLongCommandDescription(`werf helps to implement and support Continuous Integration and Continuous Delivery.`),
		SilenceUsage:  true,
		SilenceErrors: true,
	})

	groups := &templates.CommandGroups{}
	*groups = append(*groups, templates.CommandGroups{
		{
			Message: "Delivery commands",
			Commands: []*cobra.Command{
				converge.NewCmd(ctx),
				plan.NewCmd(ctx),
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
				includesCmd(ctx),
			},
		},
		{
			Message: "Low-level management commands",
			Commands: []*cobra.Command{
				configCmd(ctx),
				managedImagesCmd(ctx),
				hostCmd(ctx),
				helmCmd,
				crCmd(ctx),
				kubectl2.ReplaceKubectlDocs(kubectl.NewCmd(ctx)),
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

	return rootCmd, nil
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
		Short: "Authenticate with container registry.",
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

func includesCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "includes",
		Short: "Work with werf includes",
	})
	cmd.AddCommand(
		includes_update.NewCmd(ctx),
		includes_lsfiles.NewCmd(ctx),
		includes_getfile.NewCmd(ctx),
	)

	return cmd
}

func SetupTelemetryInit(rootCmd *cobra.Command) {
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
					telemetry.LogF("error: %s\n", err)
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

func PrintStackTraces() {
	if os.Getenv("WERF_PRINT_STACK_TRACES") != "1" {
		return
	}

	period := 5
	if prd := os.Getenv("WERF_PRINT_STACK_TRACES_PERIOD"); prd != "" {
		p, err := strconv.Atoi(prd)
		if err == nil {
			period = p
		}
	}

	go func() {
		for {
			buf := make([]byte, 1<<20)
			runtime.Stack(buf, true)
			fmt.Printf("%s", buf)

			time.Sleep(time.Second * time.Duration(period))
		}
	}()
}
