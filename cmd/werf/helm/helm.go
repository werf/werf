package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/cmd/werf/common"
	helm2 "github.com/werf/werf/cmd/werf/docs/replacers/helm"
	helm_secret_decrypt "github.com/werf/werf/cmd/werf/helm/secret/decrypt"
	helm_secret_encrypt "github.com/werf/werf/cmd/werf/helm/secret/encrypt"
	helm_secret_file_decrypt "github.com/werf/werf/cmd/werf/helm/secret/file/decrypt"
	helm_secret_file_edit "github.com/werf/werf/cmd/werf/helm/secret/file/edit"
	helm_secret_file_encrypt "github.com/werf/werf/cmd/werf/helm/secret/file/encrypt"
	helm_secret_generate_secret_key "github.com/werf/werf/cmd/werf/helm/secret/generate_secret_key"
	helm_secret_rotate_secret_key "github.com/werf/werf/cmd/werf/helm/secret/rotate_secret_key"
	helm_secret_values_decrypt "github.com/werf/werf/cmd/werf/helm/secret/values/decrypt"
	helm_secret_values_edit "github.com/werf/werf/cmd/werf/helm/secret/values/edit"
	helm_secret_values_encrypt "github.com/werf/werf/cmd/werf/helm/secret/values/encrypt"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/werf"
)

var _commonCmdData common.CmdData

func IsHelm3Mode() bool {
	return os.Getenv("WERF_HELM3_MODE") == "1"
}

func NewCmd(ctx context.Context) (*cobra.Command, error) {
	var namespace string
	ctx = common.NewContextWithCmdData(ctx, &_commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:          "helm",
		Short:        "Manage application deployment with helm",
		SilenceUsage: true,
	})

	registryClient, err := common.NewHelmRegistryClientWithoutInit(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create helm registry client: %w", err)
	}

	actionConfig := new(action.Configuration)
	actionConfig.RegistryClient = registryClient

	wc := chart_extender.NewWerfChartStub(ctx, false)

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender:               wc,
		SubchartExtenderFactoryFunc: func() chart.ChartExtender { return chart_extender.NewWerfChartStub(ctx, false) },
	}

	os.Setenv("HELM_EXPERIMENTAL_OCI", "1")

	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", *helm_v3.Settings.GetNamespaceP(), "namespace scope for this request")
	common.SetupTmpDir(&_commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&_commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupKubeConfig(&_commonCmdData, cmd)
	common.SetupKubeConfigBase64(&_commonCmdData, cmd)
	common.SetupKubeContext(&_commonCmdData, cmd)
	common.SetupStatusProgressPeriod(&_commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&_commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&_commonCmdData, cmd)
	common.SetupLogOptions(&_commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&_commonCmdData, cmd)
	common.SetupDockerConfig(&_commonCmdData, cmd, "")

	dependencyCmd := helm_v3.NewDependencyCmd(actionConfig, os.Stdout)
	for _, depCmd := range dependencyCmd.Commands() {
		if depCmd.Name() == "update" {
			oldRunE := depCmd.RunE
			depCmd.RunE = func(cmd *cobra.Command, args []string) error {
				if err := oldRunE(cmd, args); err != nil {
					return err
				}

				chartpath := "."
				if len(args) > 0 {
					chartpath = filepath.Clean(args[0])
				}

				ch, err := loader.LoadDir(chartpath)
				if err != nil {
					return fmt.Errorf("error loading chart %q: %w", chartpath, err)
				}

				return chart_extender.CopyChartDependenciesIntoCache(cmd.Context(), chartpath, ch)
			}
		}
	}

	cmd.AddCommand(
		helm2.ReplaceHelmUninstallDocs(helm_v3.NewUninstallCmd(actionConfig, os.Stdout, helm_v3.UninstallCmdOptions{
			StagesSplitter: helm.NewStagesSplitter(),
		})),
		helm2.ReplaceHelmDependencyDocs(dependencyCmd),
		helm2.ReplaceHelmGetDocs(helm_v3.NewGetCmd(actionConfig, os.Stdout)),
		helm2.ReplaceHelmHistoryDocs(helm_v3.NewHistoryCmd(actionConfig, os.Stdout)),
		NewLintCmd(actionConfig, wc),
		helm2.ReplaceHelmListDocs(helm_v3.NewListCmd(actionConfig, os.Stdout)),
		NewTemplateCmd(actionConfig, wc, &namespace),
		helm2.ReplaceHelmRepoDocs(helm_v3.NewRepoCmd(os.Stdout)),
		helm2.ReplaceHelmRollbackDocs(helm_v3.NewRollbackCmd(actionConfig, os.Stdout, helm_v3.RollbackCmdOptions{
			StagesSplitter:              helm.NewStagesSplitter(),
			StagesExternalDepsGenerator: helm.NewStagesExternalDepsGenerator(&actionConfig.RESTClientGetter, &namespace),
		})),
		NewInstallCmd(actionConfig, wc, &namespace),
		helm2.ReplaceHelmUpgradeDocs(NewUpgradeCmd(actionConfig, wc, &namespace)),
		helm2.ReplaceHelmCreateDocs(helm_v3.NewCreateCmd(os.Stdout)),
		helm2.ReplaceHelmEnvDocs(helm_v3.NewEnvCmd(os.Stdout)),
		helm2.ReplaceHelmPackageDocs(helm_v3.NewPackageCmd(actionConfig, os.Stdout)),
		helm2.ReplaceHelmPluginDocs(helm_v3.NewPluginCmd(os.Stdout)),
		helm2.ReplaceHelmPullDocs(helm_v3.NewPullCmd(actionConfig, os.Stdout)),
		helm2.ReplaceHelmSearchDocs(helm_v3.NewSearchCmd(os.Stdout)),
		helm2.ReplaceHelmShowDocs(helm_v3.NewShowCmd(actionConfig, os.Stdout)),
		helm2.ReplaceHelmStatusDocs(helm_v3.NewStatusCmd(actionConfig, os.Stdout)),
		helm_v3.NewTestCmd(actionConfig, os.Stdout),
		helm2.ReplaceHelmVerifyDocs(helm_v3.NewVerifyCmd(os.Stdout)),
		helm2.ReplaceHelmVersionDocs(helm_v3.NewVersionCmd(os.Stdout)),
		secretCmd(ctx),
		NewGetAutogeneratedValuesCmd(ctx),
		NewGetNamespaceCmd(ctx),
		NewGetReleaseCmd(ctx),
		NewMigrate2To3Cmd(ctx),
		helm_v3.NewRegistryCmd(actionConfig, os.Stdout),
	)

	if IsHelm3Mode() {
		helm_v3.LoadPlugins(cmd, os.Stdout)
	} else {
		func() {
			if len(os.Args) > 1 {
				saveArgs := os.Args
				os.Args = os.Args[1:]
				defer func() {
					os.Args = saveArgs
				}()
			}

			helm_v3.LoadPlugins(cmd, os.Stdout)
		}()
	}

	commandsQueue := []*cobra.Command{cmd}
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
				// NOTE: Common init block for all runnable commands.
				ctx := cmd.Context()

				if err := werf.Init(*_commonCmdData.TmpDir, *_commonCmdData.HomeDir); err != nil {
					return err
				}

				if err := common.ProcessLogOptions(&_commonCmdData); err != nil {
					common.PrintHelp(cmd)
					return err
				}

				os.Setenv("WERF_HELM3_MODE", "1")
				os.Setenv("HELM_NAMESPACE", namespace)

				stubCommitDate := time.Unix(0, 0)

				if vals, err := helpers.GetServiceValues(ctx, "PROJECT", "REPO", nil, helpers.ServiceValuesOptions{
					Namespace:  namespace,
					IsStub:     true,
					CommitHash: "COMMIT_HASH",
					CommitDate: &stubCommitDate,
				}); err != nil {
					return fmt.Errorf("error creating service values: %w", err)
				} else {
					wc.SetStubServiceValues(vals)
				}

				common.InitHelmRegistryClient(registryClient, *_commonCmdData.DockerConfig, *_commonCmdData.InsecureHelmDependencies)

				common.SetupOndemandKubeInitializer(*_commonCmdData.KubeContext, *_commonCmdData.KubeConfig, *_commonCmdData.KubeConfigBase64, *_commonCmdData.KubeConfigPathMergeList)

				helm.InitActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, helm_v3.Settings, actionConfig, helm.InitActionConfigOptions{
					StatusProgressPeriod:      time.Duration(*_commonCmdData.StatusProgressPeriodSeconds) * time.Second,
					HooksStatusProgressPeriod: time.Duration(*_commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
					KubeConfigOptions: kube.KubeConfigOptions{
						Context:             *_commonCmdData.KubeContext,
						ConfigPath:          *_commonCmdData.KubeConfig,
						ConfigPathMergeList: *_commonCmdData.KubeConfigPathMergeList,
						ConfigDataBase64:    *_commonCmdData.KubeConfigBase64,
					},
					ReleasesHistoryMax: *_commonCmdData.ReleasesHistoryMax,
				})

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

	if len(os.Args) > 1 {
		cmd.PersistentFlags().ParseErrorsWhitelist.UnknownFlags = true
		cmd.PersistentFlags().Parse(os.Args[1:])
	}

	return cmd, nil
}

func secretCmd(ctx context.Context) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "secret",
		Short: "Work with secrets",
	})

	fileCmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "file",
		Short: "Work with secret files",
	})

	fileCmd.AddCommand(
		helm_secret_file_encrypt.NewCmd(ctx),
		helm_secret_file_decrypt.NewCmd(ctx),
		helm_secret_file_edit.NewCmd(ctx),
	)

	valuesCmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "values",
		Short: "Work with secret values files",
	})

	valuesCmd.AddCommand(
		helm_secret_values_encrypt.NewCmd(ctx),
		helm_secret_values_decrypt.NewCmd(ctx),
		helm_secret_values_edit.NewCmd(ctx),
	)

	cmd.AddCommand(
		fileCmd,
		valuesCmd,
		helm_secret_generate_secret_key.NewCmd(ctx),
		helm_secret_encrypt.NewCmd(ctx),
		helm_secret_decrypt.NewCmd(ctx),
		helm_secret_rotate_secret_key.NewCmd(ctx),
	)

	return cmd
}
