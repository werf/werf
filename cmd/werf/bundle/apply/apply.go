package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gookit/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/kubedog/pkg/trackers/dyntracker/logstore"
	"github.com/werf/kubedog/pkg/trackers/dyntracker/statestore"
	kubeutil "github.com/werf/kubedog/pkg/trackers/dyntracker/util"
	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/chrttree"
	helmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/kubeclnt"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/nelm/pkg/opertn"
	"github.com/werf/nelm/pkg/pln"
	"github.com/werf/nelm/pkg/plnbuilder"
	"github.com/werf/nelm/pkg/plnexectr"
	"github.com/werf/nelm/pkg/reprt"
	"github.com/werf/nelm/pkg/resrc"
	"github.com/werf/nelm/pkg/resrcpatcher"
	"github.com/werf/nelm/pkg/resrcprocssr"
	"github.com/werf/nelm/pkg/rls"
	"github.com/werf/nelm/pkg/rlsdiff"
	"github.com/werf/nelm/pkg/rlshistor"
	"github.com/werf/nelm/pkg/track"
	"github.com/werf/nelm/pkg/utls"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/v2/pkg/deploy/lock_manager"
	"github.com/werf/werf/v2/pkg/deploy/secrets_manager"
	"github.com/werf/werf/v2/pkg/util"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag          string
	Timeout      int
	AutoRollback bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "apply",
		Short:                 "Apply bundle into Kubernetes",
		Long:                  common.GetLongCommandDescription(`Take latest bundle from the specified container registry using specified version tag or version mask and apply it as a helm chart into Kubernetes cluster.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runApply(ctx) })
		},
	})

	common.SetupEnvironment(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo, to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, false)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSetDockerConfigJsonValue(&commonCmdData, cmd)
	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd, false)
	common.SetupSecretValues(&commonCmdData, cmd, false)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	common.SetupSaveDeployReport(&commonCmdData, cmd)
	common.SetupDeployReportPath(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupRelease(&commonCmdData, cmd, false)
	common.SetupNamespace(&commonCmdData, cmd, false)
	common.SetupStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupHooksStatusProgressPeriod(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)

	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupDeployGraphPath(&commonCmdData, cmd)
	common.SetupRollbackGraphPath(&commonCmdData, cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Provide exact tag version or semver-based pattern, werf will install or upgrade to the latest version of the specified bundle ($WERF_TAG or latest by default)")

	defaultTimeout, err := util.GetIntEnvVar("WERF_TIMEOUT")
	if err != nil || defaultTimeout == nil {
		defaultTimeout = new(int64)
	}
	cmd.Flags().IntVarP(&cmdData.Timeout, "timeout", "t", int(*defaultTimeout), "Resources tracking timeout in seconds ($WERF_TIMEOUT by default)")

	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "auto-rollback", "R", util.GetBoolEnvironmentDefaultFalse("WERF_AUTO_ROLLBACK"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_AUTO_ROLLBACK by default)")
	cmd.Flags().BoolVarP(&cmdData.AutoRollback, "atomic", "", util.GetBoolEnvironmentDefaultFalse("WERF_ATOMIC"), "Enable auto rollback of the failed release to the previous deployed release version when current deploy process have failed ($WERF_ATOMIC by default)")

	return cmd
}

func runApply(ctx context.Context) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	registryMirrors, err := common.GetContainerRegistryMirror(ctx, &commonCmdData)
	if err != nil {
		return fmt.Errorf("get container registry mirrors: %w", err)
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData, registryMirrors); err != nil {
		return err
	}

	common.SetupOndemandKubeInitializer(*commonCmdData.KubeContext, *commonCmdData.KubeConfig, *commonCmdData.KubeConfigBase64, *commonCmdData.KubeConfigPathMergeList)
	if err := common.GetOndemandKubeInitializer().Init(ctx); err != nil {
		return err
	}

	userExtraAnnotations, err := common.GetUserExtraAnnotations(&commonCmdData)
	if err != nil {
		return err
	}

	userExtraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return err
	}

	repoAddress, err := commonCmdData.Repo.GetAddress()
	if err != nil {
		return err
	}

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *commonCmdData.DockerConfig, *commonCmdData.InsecureHelmDependencies)
	if err != nil {
		return fmt.Errorf("unable to create helm registry client: %w", err)
	}

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	namespace := common.GetNamespace(&commonCmdData)
	releaseName, err := common.GetRequiredRelease(&commonCmdData)
	if err != nil {
		return err
	}

	actionConfig := new(action.Configuration)
	if err := helm.InitActionConfig(ctx, common.GetOndemandKubeInitializer(), namespace, helm_v3.Settings, actionConfig, helm.InitActionConfigOptions{
		StatusProgressPeriod:      time.Duration(*commonCmdData.StatusProgressPeriodSeconds) * time.Second,
		HooksStatusProgressPeriod: time.Duration(*commonCmdData.HooksStatusProgressPeriodSeconds) * time.Second,
		KubeConfigOptions: kube.KubeConfigOptions{
			Context:             *commonCmdData.KubeContext,
			ConfigPath:          *commonCmdData.KubeConfig,
			ConfigDataBase64:    *commonCmdData.KubeConfigBase64,
			ConfigPathMergeList: *commonCmdData.KubeConfigPathMergeList,
		},
		ReleasesHistoryMax: *commonCmdData.ReleasesHistoryMax,
		RegistryClient:     helmRegistryClient,
	}); err != nil {
		return err
	}

	bundleTmpDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundleTmpDir)

	if err := bundles.Pull(ctx, fmt.Sprintf("%s:%s", repoAddress, cmdData.Tag), bundleTmpDir, bundlesRegistryClient); err != nil {
		return fmt.Errorf("unable to pull bundle: %w", err)
	}

	var lockManager *lock_manager.LockManager
	if m, err := lock_manager.NewLockManager(namespace); err != nil {
		return fmt.Errorf("unable to create lock manager: %w", err)
	} else {
		lockManager = m
	}

	secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{DisableSecretsDecryption: *commonCmdData.IgnoreSecretKey})

	bundle, err := chart_extender.NewBundle(ctx, bundleTmpDir, helm_v3.Settings, helmRegistryClient, secretsManager, chart_extender.BundleOptions{
		SecretValueFiles: common.GetSecretValues(&commonCmdData),
		BuildChartDependenciesOpts: command_helpers.BuildChartDependenciesOptions{
			IgnoreInvalidAnnotationsAndLabels: true,
			SkipUpdate:                        *commonCmdData.SkipDependenciesRepoRefresh,
		},
		IgnoreInvalidAnnotationsAndLabels: true,
		ExtraAnnotations:                  userExtraAnnotations,
		ExtraLabels:                       userExtraLabels,
	})
	if err != nil {
		return err
	}

	if vals, err := helpers.GetBundleServiceValues(ctx, helpers.ServiceValuesOptions{
		Env:                      *commonCmdData.Environment,
		Namespace:                namespace,
		SetDockerConfigJsonValue: *commonCmdData.SetDockerConfigJsonValue,
		DockerConfigPath:         *commonCmdData.DockerConfig,
	}); err != nil {
		return fmt.Errorf("error creating service values: %w", err)
	} else {
		bundle.SetServiceValues(vals)
	}

	loader.GlobalLoadOptions = &loader.LoadOptions{
		ChartExtender: bundle,
	}

	trackReadinessTimeout := *common.NewDuration(time.Duration(cmdData.Timeout) * time.Second)
	trackDeletionTimeout := trackReadinessTimeout
	showResourceProgress := *commonCmdData.StatusProgressPeriodSeconds != -1
	showResourceProgressPeriod := time.Duration(
		lo.Max([]int64{
			*commonCmdData.StatusProgressPeriodSeconds,
			int64(1),
		}),
	) * time.Second
	saveDeployReport := common.GetSaveDeployReport(&commonCmdData)
	deployReportPath, err := common.GetDeployReportPath(&commonCmdData)
	if err != nil {
		return fmt.Errorf("error getting deploy report path: %w", err)
	}

	deployGraphPath := common.GetDeployGraphPath(&commonCmdData)
	rollbackGraphPath := common.GetRollbackGraphPath(&commonCmdData)
	saveDeployGraph := deployGraphPath != ""
	saveRollbackGraphPath := rollbackGraphPath != ""
	networkParallelism := common.GetNetworkParallelism(&commonCmdData)

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	for key, value := range bundle.ExtraAnnotationsAndLabelsPostRenderer.ExtraAnnotations {
		if strings.HasPrefix(key, "project.werf.io/") ||
			strings.Contains(key, "ci.werf.io/") ||
			key == "werf.io/release-channel" {
			serviceAnnotations[key] = value
		} else {
			extraAnnotations[key] = value
		}
	}

	serviceAnnotations["werf.io/version"] = werf.Version
	if *commonCmdData.Environment != "" {
		serviceAnnotations["project.werf.io/env"] = *commonCmdData.Environment
	}

	extraLabels := bundle.ExtraAnnotationsAndLabelsPostRenderer.ExtraLabels

	clientFactory, err := kubeclnt.NewClientFactory()
	if err != nil {
		return fmt.Errorf("error creating kube client factory: %w", err)
	}

	releaseNamespace := resrc.NewReleaseNamespace(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": namespace,
			},
		},
	}, resrc.ReleaseNamespaceOptions{
		Mapper: clientFactory.Mapper(),
	})

	// FIXME(ilya-lesikov): there is more chartpath options, are they needed?
	chartPathOptions := action.ChartPathOptions{}
	chartPathOptions.SetRegistryClient(actionConfig.RegistryClient)

	actionConfig.Releases.MaxHistory = *commonCmdData.ReleasesHistoryMax

	return command_helpers.LockReleaseWrapper(ctx, releaseName, lockManager, func() error {
		log.Default.Info(ctx, color.Style{color.Bold, color.Green}.Render("Starting release")+" %q (namespace: %q)", releaseName, releaseNamespace.Name())

		log.Default.Info(ctx, "Constructing release history")
		history, err := rlshistor.NewHistory(releaseName, releaseNamespace.Name(), actionConfig.Releases, rlshistor.HistoryOptions{
			Mapper:          clientFactory.Mapper(),
			DiscoveryClient: clientFactory.Discovery(),
		})
		if err != nil {
			return fmt.Errorf("error constructing release history: %w", err)
		}

		prevRelease, prevReleaseFound, err := history.LastRelease()
		if err != nil {
			return fmt.Errorf("error getting last deployed release: %w", err)
		}

		prevDeployedRelease, prevDeployedReleaseFound, err := history.LastDeployedRelease()
		if err != nil {
			return fmt.Errorf("error getting last deployed release: %w", err)
		}

		var newRevision int
		var firstDeployed time.Time
		if prevReleaseFound {
			newRevision = prevRelease.Revision() + 1
			firstDeployed = prevRelease.FirstDeployed()
		} else {
			newRevision = 1
		}

		var deployType helmcommon.DeployType
		if prevReleaseFound && prevDeployedReleaseFound {
			deployType = helmcommon.DeployTypeUpgrade
		} else if prevReleaseFound {
			deployType = helmcommon.DeployTypeInstall
		} else {
			deployType = helmcommon.DeployTypeInitial
		}

		log.Default.Info(ctx, "Constructing chart tree")
		chartTree, err := chrttree.NewChartTree(
			ctx,
			bundle.Dir,
			releaseName,
			releaseNamespace.Name(),
			newRevision,
			deployType,
			actionConfig,
			chrttree.ChartTreeOptions{
				StringSetValues: common.GetSetString(&commonCmdData),
				SetValues:       common.GetSet(&commonCmdData),
				FileValues:      common.GetSetFile(&commonCmdData),
				ValuesFiles:     common.GetValues(&commonCmdData),
				Mapper:          clientFactory.Mapper(),
				DiscoveryClient: clientFactory.Discovery(),
			},
		)
		if err != nil {
			return fmt.Errorf("error constructing chart tree: %w", err)
		}

		notes := chartTree.Notes()

		var prevRelGeneralResources []*resrc.GeneralResource
		if prevReleaseFound {
			prevRelGeneralResources = prevRelease.GeneralResources()
		}

		log.Default.Info(ctx, "Processing resources")
		resProcessor := resrcprocssr.NewDeployableResourcesProcessor(
			deployType,
			releaseName,
			releaseNamespace,
			chartTree.StandaloneCRDs(),
			chartTree.HookResources(),
			chartTree.GeneralResources(),
			prevRelGeneralResources,
			resrcprocssr.DeployableResourcesProcessorOptions{
				NetworkParallelism: networkParallelism,
				ReleasableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
					resrcpatcher.NewExtraMetadataPatcher(extraAnnotations, extraLabels),
				},
				ReleasableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
					resrcpatcher.NewExtraMetadataPatcher(extraAnnotations, extraLabels),
				},
				DeployableStandaloneCRDsPatchers: []resrcpatcher.ResourcePatcher{
					resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
				},
				DeployableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
					resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
				},
				DeployableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
					resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
				},
				KubeClient:         clientFactory.KubeClient(),
				Mapper:             clientFactory.Mapper(),
				DiscoveryClient:    clientFactory.Discovery(),
				AllowClusterAccess: true,
			},
		)

		if err := resProcessor.Process(ctx); err != nil {
			return fmt.Errorf("error processing deployable resources: %w", err)
		}

		log.Default.Info(ctx, "Constructing new release")
		newRel, err := rls.NewRelease(releaseName, releaseNamespace.Name(), newRevision, chartTree.ReleaseValues(), chartTree.LegacyChart(), resProcessor.ReleasableHookResources(), resProcessor.ReleasableGeneralResources(), notes, rls.ReleaseOptions{
			FirstDeployed: firstDeployed,
			Mapper:        clientFactory.Mapper(),
		})
		if err != nil {
			return fmt.Errorf("error constructing new release: %w", err)
		}

		taskStore := statestore.NewTaskStore()
		logStore := kubeutil.NewConcurrent(
			logstore.NewLogStore(),
		)

		log.Default.Info(ctx, "Constructing new deploy plan")
		deployPlanBuilder := plnbuilder.NewDeployPlanBuilder(
			deployType,
			taskStore,
			logStore,
			resProcessor.DeployableReleaseNamespaceInfo(),
			resProcessor.DeployableStandaloneCRDsInfos(),
			resProcessor.DeployableHookResourcesInfos(),
			resProcessor.DeployableGeneralResourcesInfos(),
			resProcessor.DeployablePrevReleaseGeneralResourcesInfos(),
			newRel,
			history,
			clientFactory.KubeClient(),
			clientFactory.Static(),
			clientFactory.Dynamic(),
			clientFactory.Discovery(),
			clientFactory.Mapper(),
			plnbuilder.DeployPlanBuilderOptions{
				PrevRelease:         prevRelease,
				PrevDeployedRelease: prevDeployedRelease,
				CreationTimeout:     trackReadinessTimeout,
				ReadinessTimeout:    trackReadinessTimeout,
				DeletionTimeout:     trackDeletionTimeout,
			},
		)

		plan, err := deployPlanBuilder.Build(ctx)
		if err != nil {
			if deployGraphPath == "" {
				if file, err := os.CreateTemp("", "werf-deploy-plan-*.dot"); err != nil {
					log.Default.Error(ctx, "Error creating temporary file for deploy graph: %s", err)
					return fmt.Errorf("error building deploy plan: %w", err)
				} else {
					deployGraphPath = file.Name()
				}
			}

			if err := plan.SaveDOT(deployGraphPath); err != nil {
				log.Default.Error(ctx, "Error saving deploy graph: %s", err)
			}
			log.Default.Warn(ctx, "Deploy graph saved to %q for debugging", deployGraphPath)

			return fmt.Errorf("error building deploy plan: %w", err)
		}

		if saveDeployGraph {
			if err := plan.SaveDOT(deployGraphPath); err != nil {
				return fmt.Errorf("error saving deploy graph: %w", err)
			}
		}

		var releaseUpToDate bool
		if prevReleaseFound {
			releaseUpToDate, err = rlsdiff.ReleaseUpToDate(prevRelease, newRel)
			if err != nil {
				return fmt.Errorf("error checking if release is up to date: %w", err)
			}
		}

		planUseless, err := plan.Useless()
		if err != nil {
			return fmt.Errorf("error checking if deploy plan will do nothing useful: %w", err)
		}

		if releaseUpToDate && planUseless {
			if saveDeployReport {
				newRel.Skip()

				report := reprt.NewReport(
					nil,
					nil,
					nil,
					newRel,
				)

				if err := report.Save(deployReportPath); err != nil {
					log.Default.Error(ctx, "Error saving deploy report: %s", err)
				}
			}

			printNotes(ctx, notes)

			log.Default.Info(ctx, color.Style{color.Bold, color.Green}.Render(fmt.Sprintf("Skipped release %q (namespace: %q): cluster resources already as desired", releaseName, releaseNamespace.Name())))

			return nil
		}

		colorize := *commonCmdData.LogColorMode != "off"
		tablesBuilder := track.NewTablesBuilder(
			taskStore,
			logStore,
			track.TablesBuilderOptions{
				DefaultNamespace: releaseNamespace.Name(),
				Colorize:         colorize,
			},
		)

		log.Default.Info(ctx, "Starting tracking")
		stdoutTrackerStopCh := make(chan bool)
		stdoutTrackerFinishedCh := make(chan bool)

		if showResourceProgress {
			go func() {
				ticker := time.NewTicker(showResourceProgressPeriod)
				defer func() {
					ticker.Stop()
					stdoutTrackerFinishedCh <- true
				}()

				for {
					select {
					case <-ticker.C:
						printTables(ctx, tablesBuilder)
					case <-stdoutTrackerStopCh:
						printTables(ctx, tablesBuilder)
						return
					}
				}
			}()
		}

		log.Default.Info(ctx, "Executing deploy plan")
		planExecutor := plnexectr.NewPlanExecutor(plan, plnexectr.PlanExecutorOptions{
			NetworkParallelism: networkParallelism,
		})

		var criticalErrs, nonCriticalErrs []error

		planExecutionErr := planExecutor.Execute(ctx)
		if planExecutionErr != nil {
			criticalErrs = append(criticalErrs, fmt.Errorf("error executing deploy plan: %w", planExecutionErr))
		}

		var worthyCompletedOps []opertn.Operation
		if ops, found, err := plan.WorthyCompletedOperations(); err != nil {
			nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy completed operations: %w", err))
		} else if found {
			worthyCompletedOps = ops
		}

		var worthyCanceledOps []opertn.Operation
		if ops, found, err := plan.WorthyCanceledOperations(); err != nil {
			nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy canceled operations: %w", err))
		} else if found {
			worthyCanceledOps = ops
		}

		var worthyFailedOps []opertn.Operation
		if ops, found, err := plan.WorthyFailedOperations(); err != nil {
			nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy failed operations: %w", err))
		} else if found {
			worthyFailedOps = ops
		}

		var pendingReleaseCreated bool
		if ops, found, err := plan.OperationsMatch(regexp.MustCompile(fmt.Sprintf(`^%s/%s$`, opertn.TypeCreatePendingReleaseOperation, newRel.ID()))); err != nil {
			nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting pending release operation: %w", err))
		} else if !found {
			panic("no pending release operation found")
		} else {
			pendingReleaseCreated = ops[0].Status() == opertn.StatusCompleted
		}

		if planExecutionErr != nil && pendingReleaseCreated {
			wcompops, wfailops, wcancops, criterrs, noncriterrs := runFailureDeployPlan(
				ctx,
				plan,
				taskStore,
				resProcessor,
				newRel,
				prevRelease,
				history,
				clientFactory,
				networkParallelism,
			)
			worthyCompletedOps = append(worthyCompletedOps, wcompops...)
			worthyFailedOps = append(worthyFailedOps, wfailops...)
			worthyCanceledOps = append(worthyCanceledOps, wcancops...)
			criticalErrs = append(criticalErrs, criterrs...)
			nonCriticalErrs = append(nonCriticalErrs, noncriterrs...)

			if cmdData.AutoRollback && prevDeployedReleaseFound {
				wcompops, wfailops, wcancops, notes, criterrs, noncriterrs = runRollbackPlan(
					ctx,
					taskStore,
					logStore,
					releaseName,
					releaseNamespace,
					newRel,
					prevDeployedRelease,
					newRevision,
					history,
					clientFactory,
					extraAnnotations,
					serviceAnnotations,
					extraLabels,
					trackReadinessTimeout,
					trackReadinessTimeout,
					trackDeletionTimeout,
					saveRollbackGraphPath,
					rollbackGraphPath,
					networkParallelism,
				)
				worthyCompletedOps = append(worthyCompletedOps, wcompops...)
				worthyFailedOps = append(worthyFailedOps, wfailops...)
				worthyCanceledOps = append(worthyCanceledOps, wcancops...)
				criticalErrs = append(criticalErrs, criterrs...)
				nonCriticalErrs = append(nonCriticalErrs, noncriterrs...)
			}
		}

		if showResourceProgress {
			stdoutTrackerStopCh <- true
			<-stdoutTrackerFinishedCh
		}

		report := reprt.NewReport(
			worthyCompletedOps,
			worthyCanceledOps,
			worthyFailedOps,
			newRel,
		)

		report.Print(ctx)

		if saveDeployReport {
			if err := report.Save(deployReportPath); err != nil {
				nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error saving deploy report: %w", err))
			}
		}

		if len(criticalErrs) == 0 {
			printNotes(ctx, notes)
		}

		if len(criticalErrs) > 0 {
			return utls.Multierrorf("failed release %q (namespace: %q)", append(criticalErrs, nonCriticalErrs...), releaseName, releaseNamespace.Name())
		} else if len(nonCriticalErrs) > 0 {
			return utls.Multierrorf("succeeded release %q (namespace: %q), but non-critical errors encountered", nonCriticalErrs, releaseName, releaseNamespace.Name())
		} else {
			log.Default.Info(ctx, color.Style{color.Bold, color.Green}.Render(fmt.Sprintf("Succeeded release %q (namespace: %q)", releaseName, releaseNamespace.Name())))
			return nil
		}
	})
}

func runFailureDeployPlan(ctx context.Context, failedPlan *pln.Plan, taskStore *statestore.TaskStore, resProcessor *resrcprocssr.DeployableResourcesProcessor, newRel, prevRelease *rls.Release, history *rlshistor.History, clientFactory *kubeclnt.ClientFactory, networkParallelism int) (worthyCompletedOps, worthyFailedOps, worthyCanceledOps []opertn.Operation, criticalErrs, nonCriticalErrs []error) {
	log.Default.Info(ctx, "Building failure deploy plan")
	failurePlanBuilder := plnbuilder.NewDeployFailurePlanBuilder(
		failedPlan,
		taskStore,
		resProcessor.DeployableReleaseNamespaceInfo(),
		resProcessor.DeployableHookResourcesInfos(),
		resProcessor.DeployableGeneralResourcesInfos(),
		newRel,
		history,
		clientFactory.KubeClient(),
		clientFactory.Dynamic(),
		clientFactory.Mapper(),
		plnbuilder.DeployFailurePlanBuilderOptions{
			PrevRelease: prevRelease,
		},
	)

	failurePlan, err := failurePlanBuilder.Build(ctx)
	if err != nil {
		return nil, nil, nil, []error{fmt.Errorf("error building failure plan: %w", err)}, nil
	}

	if useless, err := failurePlan.Useless(); err != nil {
		return nil, nil, nil, []error{fmt.Errorf("error checking if failure plan will do nothing useful: %w", err)}, nil
	} else if useless {
		return nil, nil, nil, nil, nil
	}

	log.Default.Info(ctx, "Executing failure deploy plan")
	failurePlanExecutor := plnexectr.NewPlanExecutor(failurePlan, plnexectr.PlanExecutorOptions{
		NetworkParallelism: networkParallelism,
	})

	if err := failurePlanExecutor.Execute(ctx); err != nil {
		criticalErrs = append(criticalErrs, fmt.Errorf("error executing failure plan: %w", err))
	}

	if ops, found, err := failurePlan.WorthyCompletedOperations(); err != nil {
		nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy completed operations: %w", err))
	} else if found {
		worthyCompletedOps = append(worthyCompletedOps, ops...)
	}

	if ops, found, err := failurePlan.WorthyFailedOperations(); err != nil {
		nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy failed operations: %w", err))
	} else if found {
		worthyFailedOps = append(worthyFailedOps, ops...)
	}

	if ops, found, err := failurePlan.WorthyCanceledOperations(); err != nil {
		nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy canceled operations: %w", err))
	} else if found {
		worthyCanceledOps = append(worthyCanceledOps, ops...)
	}

	return worthyCompletedOps, worthyFailedOps, worthyCanceledOps, criticalErrs, nonCriticalErrs
}

func runRollbackPlan(
	ctx context.Context,
	taskStore *statestore.TaskStore,
	logStore *kubeutil.Concurrent[*logstore.LogStore],
	releaseName string,
	releaseNamespace *resrc.ReleaseNamespace,
	failedRelease, prevDeployedRelease *rls.Release,
	failedRevision int,
	history *rlshistor.History,
	clientFactory *kubeclnt.ClientFactory,
	extraAnnotations, serviceAnnotations, extraLabels map[string]string,
	trackReadinessTimeout, trackCreationTimeout, trackDeletionTimeout time.Duration,
	saveRollbackGraph bool,
	rollbackGraphPath string,
	networkParallelism int,
) (
	worthyCompletedOps, worthyFailedOps, worthyCanceledOps []opertn.Operation, notes string,
	criticalErrs, nonCriticalErrs []error,
) {
	log.Default.Info(ctx, "Processing rollback resources")
	resProcessor := resrcprocssr.NewDeployableResourcesProcessor(
		helmcommon.DeployTypeRollback,
		releaseName,
		releaseNamespace,
		nil,
		prevDeployedRelease.HookResources(),
		prevDeployedRelease.GeneralResources(),
		failedRelease.GeneralResources(),
		resrcprocssr.DeployableResourcesProcessorOptions{
			NetworkParallelism: networkParallelism,
			ReleasableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
				resrcpatcher.NewExtraMetadataPatcher(extraAnnotations, extraLabels),
			},
			ReleasableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
				resrcpatcher.NewExtraMetadataPatcher(extraAnnotations, extraLabels),
			},
			DeployableStandaloneCRDsPatchers: []resrcpatcher.ResourcePatcher{
				resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
			},
			DeployableHookResourcePatchers: []resrcpatcher.ResourcePatcher{
				resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
			},
			DeployableGeneralResourcePatchers: []resrcpatcher.ResourcePatcher{
				resrcpatcher.NewExtraMetadataPatcher(lo.Assign(extraAnnotations, serviceAnnotations), extraLabels),
			},
			KubeClient:         clientFactory.KubeClient(),
			Mapper:             clientFactory.Mapper(),
			DiscoveryClient:    clientFactory.Discovery(),
			AllowClusterAccess: true,
		},
	)

	if err := resProcessor.Process(ctx); err != nil {
		return nil, nil, nil, "", []error{fmt.Errorf("error processing rollback resources: %w", err)}, nonCriticalErrs
	}

	rollbackRevision := failedRevision + 1

	log.Default.Info(ctx, "Constructing rollback release")
	rollbackRel, err := rls.NewRelease(
		releaseName,
		releaseNamespace.Name(),
		rollbackRevision,
		prevDeployedRelease.Values(),
		prevDeployedRelease.LegacyChart(),
		resProcessor.ReleasableHookResources(),
		resProcessor.ReleasableGeneralResources(),
		prevDeployedRelease.Notes(),
		rls.ReleaseOptions{
			FirstDeployed: prevDeployedRelease.FirstDeployed(),
			Mapper:        clientFactory.Mapper(),
		},
	)
	if err != nil {
		return nil, nil, nil, "", []error{fmt.Errorf("error constructing rollback release: %w", err)}, nonCriticalErrs
	}

	log.Default.Info(ctx, "Constructing rollback plan")
	rollbackPlanBuilder := plnbuilder.NewDeployPlanBuilder(
		helmcommon.DeployTypeRollback,
		taskStore,
		logStore,
		resProcessor.DeployableReleaseNamespaceInfo(),
		nil,
		resProcessor.DeployableHookResourcesInfos(),
		resProcessor.DeployableGeneralResourcesInfos(),
		resProcessor.DeployablePrevReleaseGeneralResourcesInfos(),
		rollbackRel,
		history,
		clientFactory.KubeClient(),
		clientFactory.Static(),
		clientFactory.Dynamic(),
		clientFactory.Discovery(),
		clientFactory.Mapper(),
		plnbuilder.DeployPlanBuilderOptions{
			PrevRelease:         failedRelease,
			PrevDeployedRelease: prevDeployedRelease,
			CreationTimeout:     trackCreationTimeout,
			ReadinessTimeout:    trackReadinessTimeout,
			DeletionTimeout:     trackDeletionTimeout,
		},
	)

	rollbackPlan, err := rollbackPlanBuilder.Build(ctx)
	if err != nil {
		return nil, nil, nil, "", []error{fmt.Errorf("error building rollback plan: %w", err)}, nonCriticalErrs
	}

	if saveRollbackGraph {
		if err := rollbackPlan.SaveDOT(rollbackGraphPath); err != nil {
			nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error saving rollback graph: %w", err))
		}
	}

	if useless, err := rollbackPlan.Useless(); err != nil {
		return nil, nil, nil, "", []error{fmt.Errorf("error checking if rollback plan will do nothing useful: %w", err)}, nonCriticalErrs
	} else if useless {
		log.Default.Info(ctx, color.Style{color.Bold, color.Green}.Render("Skipped rollback release")+" %q (namespace: %q): cluster resources already as desired", releaseName, releaseNamespace.Name())
		return nil, nil, nil, "", criticalErrs, nonCriticalErrs
	}

	log.Default.Info(ctx, "Executing rollback plan")
	rollbackPlanExecutor := plnexectr.NewPlanExecutor(rollbackPlan, plnexectr.PlanExecutorOptions{
		NetworkParallelism: networkParallelism,
	})

	rollbackPlanExecutionErr := rollbackPlanExecutor.Execute(ctx)
	if rollbackPlanExecutionErr != nil {
		criticalErrs = append(criticalErrs, fmt.Errorf("error executing rollback plan: %w", err))
	}

	if ops, found, err := rollbackPlan.WorthyCompletedOperations(); err != nil {
		nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy completed operations: %w", err))
	} else if found {
		worthyCompletedOps = ops
	}

	if ops, found, err := rollbackPlan.WorthyFailedOperations(); err != nil {
		nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy failed operations: %w", err))
	} else if found {
		worthyFailedOps = ops
	}

	if ops, found, err := rollbackPlan.WorthyCanceledOperations(); err != nil {
		nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting worthy canceled operations: %w", err))
	} else if found {
		worthyCanceledOps = ops
	}

	var pendingRollbackReleaseCreated bool
	if ops, found, err := rollbackPlan.OperationsMatch(regexp.MustCompile(fmt.Sprintf(`^%s/%s$`, opertn.TypeCreatePendingReleaseOperation, rollbackRel.ID()))); err != nil {
		nonCriticalErrs = append(nonCriticalErrs, fmt.Errorf("error getting pending rollback release operation: %w", err))
	} else if !found {
		panic("no pending rollback release operation found")
	} else {
		pendingRollbackReleaseCreated = ops[0].Status() == opertn.StatusCompleted
	}

	if rollbackPlanExecutionErr != nil && pendingRollbackReleaseCreated {
		wcompops, wfailops, wcancops, criterrs, noncriterrs := runFailureDeployPlan(
			ctx,
			rollbackPlan,
			taskStore,
			resProcessor,
			rollbackRel,
			failedRelease,
			history,
			clientFactory,
			networkParallelism,
		)
		worthyCompletedOps = append(worthyCompletedOps, wcompops...)
		worthyFailedOps = append(worthyFailedOps, wfailops...)
		worthyCanceledOps = append(worthyCanceledOps, wcancops...)
		criticalErrs = append(criticalErrs, criterrs...)
		nonCriticalErrs = append(nonCriticalErrs, noncriterrs...)
	}

	return worthyCompletedOps, worthyFailedOps, worthyCanceledOps, rollbackRel.Notes(), criticalErrs, nonCriticalErrs
}

func printTables(ctx context.Context, tablesBuilder *track.TablesBuilder) {
	maxTableWidth := logboek.Context(ctx).Streams().ContentWidth() - 2
	tablesBuilder.SetMaxTableWidth(maxTableWidth)

	if tables, nonEmpty := tablesBuilder.BuildEventTables(); nonEmpty {
		headers := lo.Keys(tables)
		sort.Strings(headers)

		for _, header := range headers {
			logboek.Context(ctx).LogBlock(header).Do(func() {
				tables[header].SuppressTrailingSpaces()
				logboek.Context(ctx).LogLn(tables[header].Render())
			})
		}
	}

	if tables, nonEmpty := tablesBuilder.BuildLogTables(); nonEmpty {
		headers := lo.Keys(tables)
		sort.Strings(headers)

		for _, header := range headers {
			logboek.Context(ctx).LogBlock(header).Do(func() {
				tables[header].SuppressTrailingSpaces()
				logboek.Context(ctx).LogLn(tables[header].Render())
			})
		}
	}

	if table, nonEmpty := tablesBuilder.BuildProgressTable(); nonEmpty {
		logboek.Context(ctx).LogBlock(color.Style{color.Bold, color.Blue}.Render("Progress status")).Do(func() {
			table.SuppressTrailingSpaces()
			logboek.Context(ctx).LogLn(table.Render())
		})
	}
}

func printNotes(ctx context.Context, notes string) {
	if notes == "" {
		return
	}

	log.Default.InfoBlock(ctx, color.Style{color.Bold, color.Blue}.Render("Release notes")).Do(func() {
		log.Default.Info(ctx, notes)
	})
}
