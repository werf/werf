package rollback

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/nelm/pkg/action"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/config/deploy_params"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Revision int
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)

	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the Helm release",
		Long:  common.GetLongCommandDescription(GetRollbackDocs().Long),
		Example: `# Rollback the Helm release to the specified revision
werf rollback --revision 10`,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.DocsLongMD: GetRollbackDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return run(ctx)
			})
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupAllowIncludesUpdate(cmd)

	lo.Must0(common.SetupKubeConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupTrackingFlags(&commonCmdData, cmd))

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	common.SetupExtraRuntimeAnnotations(&commonCmdData, cmd)
	common.SetupExtraRuntimeLabels(&commonCmdData, cmd)
	common.SetupForceAdoption(&commonCmdData, cmd)
	common.SetupNamespace(&commonCmdData, cmd, true)
	common.SetupNetworkParallelism(&commonCmdData, cmd)
	common.SetupNoRemoveManualChanges(&commonCmdData, cmd)
	common.SetupNoShowNotes(&commonCmdData, cmd)
	common.SetupRelease(&commonCmdData, cmd, true)
	common.SetupReleaseInfoAnnotations(&commonCmdData, cmd)
	common.SetupReleaseLabel(&commonCmdData, cmd)
	common.SetupReleaseStorageDriver(&commonCmdData, cmd)
	common.SetupReleaseStorageSQLConnection(&commonCmdData, cmd)
	common.SetupReleasesHistoryMax(&commonCmdData, cmd)
	common.SetupRollbackGraphPath(&commonCmdData, cmd)
	common.SetupRollbackReportPath(&commonCmdData, cmd)
	common.SetupSaveRollbackReport(&commonCmdData, cmd)

	cmd.Flags().IntVarP(&cmdData.Revision, "revision", "", 0, "Revision number to rollback to (if not specified, rolls back to previous revision)")

	return cmd
}

func run(ctx context.Context) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)

	_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitWerf:           true,
		InitGitDataManager: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	var gitNotFoundErr *common.GitWorktreeNotFoundError
	if err != nil {
		if !errors.As(err, &gitNotFoundErr) {
			return fmt.Errorf("get giterminism manager: %w", err)
		}
	}

	releaseNamespace, releaseName, projectName, err := getNamespaceAndRelease(ctx, gitNotFoundErr == nil, giterminismManager)
	if err != nil {
		return fmt.Errorf("get release name and namespace: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	if annos, err := common.GetUserExtraAnnotations(&commonCmdData); err != nil {
		return fmt.Errorf("get user extra annotations: %w", err)
	} else {
		for key, value := range annos {
			if strings.HasPrefix(key, "project.werf.io/") ||
				strings.Contains(key, "ci.werf.io/") ||
				key == "werf.io/release-channel" {
				serviceAnnotations[key] = value
			} else {
				extraAnnotations[key] = value
			}
		}
	}

	serviceAnnotations["werf.io/version"] = werf.Version
	serviceAnnotations["project.werf.io/name"] = projectName
	serviceAnnotations["project.werf.io/env"] = commonCmdData.Environment

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
	}

	extraRuntimeAnnotations := lo.Assign(commonCmdData.ExtraRuntimeAnnotations, extraAnnotations, serviceAnnotations)
	extraRuntimeLabels := lo.Assign(commonCmdData.ExtraRuntimeLabels, extraLabels)
	releaseInfoAnnotations := lo.Assign(commonCmdData.ReleaseInfoAnnotations, serviceAnnotations)

	var rollbackReportPath string
	if commonCmdData.SaveRollbackReport {
		rollbackReportPath = commonCmdData.RollbackReportPath
	}

	releaseLabels, err := common.GetReleaseLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get release labels: %w", err)
	}

	ctx = log.SetupLogging(ctx, cmp.Or(common.GetNelmLogLevel(&commonCmdData), action.DefaultReleaseRollbackLogLevel), log.SetupLoggingOptions{
		ColorMode: *commonCmdData.LogColorMode,
	})

	if err := action.ReleaseRollback(ctx, releaseName, releaseNamespace, action.ReleaseRollbackOptions{
		ExtraRuntimeAnnotations:     extraRuntimeAnnotations,
		ExtraRuntimeLabels:          extraRuntimeLabels,
		ForceAdoption:               commonCmdData.ForceAdoption,
		KubeConnectionOptions:       commonCmdData.KubeConnectionOptions,
		NetworkParallelism:          commonCmdData.NetworkParallelism,
		NoRemoveManualChanges:       commonCmdData.NoRemoveManualChanges,
		NoShowNotes:                 commonCmdData.NoShowNotes,
		ReleaseHistoryLimit:         commonCmdData.ReleaseHistoryLimit,
		ReleaseInfoAnnotations:      releaseInfoAnnotations,
		ReleaseLabels:               releaseLabels,
		ReleaseStorageDriver:        commonCmdData.ReleaseStorageDriver,
		ReleaseStorageSQLConnection: commonCmdData.ReleaseStorageSQLConnection,
		Revision:                    cmdData.Revision,
		RollbackGraphPath:           commonCmdData.RollbackGraphPath,
		RollbackReportPath:          rollbackReportPath,
		TrackingOptions:             commonCmdData.TrackingOptions,
	}); err != nil {
		return fmt.Errorf("release install: %w", err)
	}

	return nil
}

func getNamespaceAndRelease(ctx context.Context, gitFound bool, giterminismMgr giterminism_manager.Interface) (string, string, string, error) {
	namespaceSpecified := commonCmdData.Namespace != ""
	releaseSpecified := commonCmdData.Release != ""

	var namespace string
	var release string
	var project string
	if namespaceSpecified && releaseSpecified {
		namespace = commonCmdData.Namespace
		release = commonCmdData.Release
	} else if gitFound {
		common.ProcessLogProjectDir(&commonCmdData, giterminismMgr.ProjectDir())

		_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismMgr, common.GetWerfConfigOptions(&commonCmdData, true))
		if err != nil {
			return "", "", "", fmt.Errorf("unable to load werf config: %w", err)
		}
		logboek.LogOptionalLn()

		project = werfConfig.Meta.Project

		if namespaceSpecified {
			namespace = commonCmdData.Namespace
		} else {
			namespace, err = deploy_params.GetKubernetesNamespace(commonCmdData.Namespace, commonCmdData.Environment, werfConfig)
			if err != nil {
				return "", "", "", err
			}
		}

		if releaseSpecified {
			release = commonCmdData.Release
		} else {
			release, err = deploy_params.GetHelmRelease(commonCmdData.Release, commonCmdData.Environment, namespace, werfConfig)
			if err != nil {
				return "", "", "", err
			}
		}
	} else {
		return "", "", "", fmt.Errorf("no git with werf project found: rollback should either be executed in a git repository, or with --namespace and --release specified")
	}

	return namespace, release, project, nil
}
