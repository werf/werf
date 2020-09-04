package helm_v3

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/deploy/helm_v3"

	"helm.sh/helm/v3/pkg/action"

	"github.com/werf/werf/cmd/werf/common"
	cmd_werf_common "github.com/werf/werf/cmd/werf/common"

	"github.com/spf13/cobra"

	cmd_helm "helm.sh/helm/v3/cmd/helm"
)

var commonCmdData cmd_werf_common.CmdData

func NewCmd() *cobra.Command {
	actionConfig := new(action.Configuration)

	cmd := &cobra.Command{
		Use:   "helm-v3",
		Short: "Manage application deployment with helm",
	}

	cmd.PersistentFlags().StringVarP(cmd_helm.Settings.GetNamespaceP(), "namespace", "n", *cmd_helm.Settings.GetNamespaceP(), "namespace scope for this request")
	cmd_werf_common.SetupKubeConfig(&commonCmdData, cmd)
	cmd_werf_common.SetupKubeConfigBase64(&commonCmdData, cmd)
	cmd_werf_common.SetupKubeContext(&commonCmdData, cmd)
	cmd_werf_common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	cmd_werf_common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)

	cmd.AddCommand(
		cmd_helm.NewUninstallCmd(actionConfig, os.Stdout),
		cmd_helm.NewDependencyCmd(os.Stdout),
		cmd_helm.NewGetCmd(actionConfig, os.Stdout),
		cmd_helm.NewHistoryCmd(actionConfig, os.Stdout),
		cmd_helm.NewLintCmd(os.Stdout),
		cmd_helm.NewListCmd(actionConfig, os.Stdout),
		cmd_helm.NewTemplateCmd(actionConfig, os.Stdout),
		cmd_helm.NewRepoCmd(os.Stdout),
		cmd_helm.NewRollbackCmd(actionConfig, os.Stdout),
		NewInstallCmd(actionConfig, &commonCmdData),
		NewUpgradeCmd(actionConfig, &commonCmdData),
		cmd_helm.NewCreateCmd(os.Stdout),
		cmd_helm.NewEnvCmd(os.Stdout),
		cmd_helm.NewPackageCmd(os.Stdout),
		cmd_helm.NewPluginCmd(os.Stdout),
		cmd_helm.NewPullCmd(os.Stdout),
		cmd_helm.NewSearchCmd(os.Stdout),
		cmd_helm.NewShowCmd(os.Stdout),
		cmd_helm.NewStatusCmd(actionConfig, os.Stdout),
		cmd_helm.NewTestCmd(actionConfig, os.Stdout),
		cmd_helm.NewVerifyCmd(os.Stdout),
		cmd_helm.NewVersionCmd(os.Stdout),
	)

	cmd_helm.LoadPlugins(cmd, os.Stdout)

	commandsQueue := []*cobra.Command{cmd}
	for len(commandsQueue) > 0 {
		cmd := commandsQueue[0]
		commandsQueue = commandsQueue[1:]

		for _, cmd := range cmd.Commands() {
			commandsQueue = append(commandsQueue, cmd)
		}

		if cmd.Runnable() {
			oldRunE := cmd.RunE
			oldRun := cmd.Run

			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// NOTE: Common init block for all runnable commands.

				// FIXME: setup namespace env var for helm diff plugin

				os.Setenv("WERF_HELM3_MODE", "1")

				ctx := common.BackgroundContext()

				helm_v3.InitActionConfig(ctx, cmd_helm.Settings, actionConfig, helm_v3.InitActionConfigOptions{
					StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
					HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
				})

				if err := kube.Init(kube.InitOptions{kube.KubeConfigOptions{
					Context:          *commonCmdData.KubeContext,
					ConfigPath:       *commonCmdData.KubeConfig,
					ConfigDataBase64: *commonCmdData.KubeConfigBase64,
				}}); err != nil {
					return fmt.Errorf("cannot initialize kube: %s", err)
				}

				if err := common.InitKubedog(ctx); err != nil {
					return fmt.Errorf("cannot init kubedog: %s", err)
				}

				if oldRun != nil {
					oldRun(cmd, args)
					return nil
				} else {
					if err := oldRunE(cmd, args); err != nil {
						errValue := reflect.ValueOf(err)
						if errValue.Kind() == reflect.Struct {
							codeValue := errValue.FieldByName("code")
							if !codeValue.IsZero() {
								os.Exit(int(codeValue.Int()))
							}
						}

						return err
					}

					return nil
				}
			}
		}
	}

	return cmd
}
