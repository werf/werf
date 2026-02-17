package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/engine"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/featgate"
	"github.com/werf/nelm/pkg/log"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/util/option"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

const (
	CleaningCommandsForceOptionDescription = "First remove containers that use werf docker images which are going to be deleted"
	StubRepoAddress                        = "stub/repository"
	StubTag                                = "TAG"

	DefaultSaveBuildReport     = false
	DefaultBuildReportPathJSON = ".werf-build-report.json"
	DefaultUseBuildReport      = false

	DefaultSaveDeployReport        = false
	DefaultSaveRollbackReport      = false
	DefaultUseDeployReport         = false
	DefaultDeployReportPathJSON    = ".werf-deploy-report.json"
	DefaultRollbackReportPathJSON  = ".werf-rollback-report.json"
	DefaultUninstallReportPathJSON = ".werf-uninstall-report.json"
	DefaultSaveUninstallReport     = false
	TemplateErrHint                = "Use --debug --debug-templates (or $WERF_DEBUG and $WERF_DEBUG_TEMPLATES environment variables) to get more details about this error."
)

func init() {
	loader.NoChartLockWarning = `Cannot automatically download chart dependencies without .helm/Chart.lock or .helm/requirements.lock. Run "werf helm dependency update .helm" and commit resulting .helm/Chart.lock or .helm/requirements.lock. Committing .tgz files in .helm/charts is not required, better add "/.helm/charts/*.tgz" to the .gitignore.`
	engine.TemplateErrHint = TemplateErrHint
}

type GitWorktreeNotFoundError struct{}

func (e *GitWorktreeNotFoundError) Error() string {
	return fmt.Sprintf("werf requires a git work tree for the project to exist: unable to find a valid .git in the current directory %q or parent directories, you may also specify git work tree explicitly with --git-work-tree option (or WERF_GIT_WORK_TREE env var)", util.GetAbsoluteFilepath("."))
}

func GetLongCommandDescription(text string) string {
	return logboek.FitText(text, types.FitTextOptions{MaxWidth: 100})
}

func SetupSetDockerConfigJsonValue(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SetDockerConfigJsonValue = new(bool)
	cmd.Flags().BoolVarP(cmdData.SetDockerConfigJsonValue, "set-docker-config-json-value", "", util.GetBoolEnvironmentDefaultFalse("WERF_SET_DOCKER_CONFIG_JSON_VALUE"), "Shortcut to set current docker config into the .Values.dockerconfigjson")
}

func SetupGitWorkTree(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.GitWorkTree = new(string)
	cmd.Flags().StringVarP(cmdData.GitWorkTree, "git-work-tree", "", os.Getenv("WERF_GIT_WORK_TREE"), "Use specified git work tree dir (default $WERF_WORK_TREE or lookup for directory that contains .git in the current or parent directories)")
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.Flags().StringVarP(cmdData.Dir, "dir", "", os.Getenv("WERF_DIR"), "Use specified project directory where project’s werf.yaml and other configuration files should reside (default $WERF_DIR or current working directory)")
}

func SetupConfigPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigPath = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigPath, "config", "", os.Getenv("WERF_CONFIG"), `Use custom configuration file (default $WERF_CONFIG or werf.yaml in the project directory)`)
}

func SetupGiterminismConfigPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.GiterminismConfigRelPath = new(string)
	cmd.Flags().StringVarP(cmdData.GiterminismConfigRelPath, "giterminism-config", "", os.Getenv("WERF_GITERMINISM_CONFIG"), "Custom path to the giterminism configuration file relative to working directory (default $WERF_GITERMINISM_CONFIG or werf-giterminism.yaml in working directory)")
}

func SetupConfigTemplatesDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigTemplatesDir = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigTemplatesDir, "config-templates-dir", "", os.Getenv("WERF_CONFIG_TEMPLATES_DIR"), `Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf in working directory)`)
}

func SetupConfigRenderPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigRenderPath = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigRenderPath, "config-render-path", "", "", `Custom path for storing rendered configuration file`)
}

type SetupTmpDirOptions struct {
	Persistent bool
}

func SetupTmpDir(cmdData *CmdData, cmd *cobra.Command, opts SetupTmpDirOptions) {
	cmdData.TmpDir = new(string)
	getFlags(cmd, opts.Persistent).StringVarP(cmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)")
}

func SetupGiterminismOptions(cmdData *CmdData, cmd *cobra.Command) {
	setupLooseGiterminism(cmdData, cmd)
	setupDev(cmdData, cmd)
	setupDevIgnore(cmdData, cmd)
	setupDevBranch(cmdData, cmd)
}

func setupLooseGiterminism(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LooseGiterminism = new(bool)
	cmd.Flags().BoolVarP(cmdData.LooseGiterminism, "loose-giterminism", "", util.GetBoolEnvironmentDefaultFalse("WERF_LOOSE_GITERMINISM"), "Loose werf giterminism mode restrictions")
}

func setupDev(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dev = new(bool)
	cmd.Flags().BoolVarP(cmdData.Dev, "dev", "", util.GetBoolEnvironmentDefaultFalse("WERF_DEV"), `Enable development mode (default $WERF_DEV).
The mode allows working with project files without doing redundant commits during debugging and development`)
}

func setupDevIgnore(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DevIgnore = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.DevIgnore, "dev-ignore", "", []string{}, `Add rules to ignore tracked and untracked changes in development mode (can specify multiple).
Also, can be specified with $WERF_DEV_IGNORE_* (e.g. $WERF_DEV_IGNORE_TESTS=*_test.go, $WERF_DEV_IGNORE_DOCS=path/to/docs)`)
}

func setupDevBranch(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DevBranch = new(string)

	defaultValue := "_werf-dev"
	envValue := os.Getenv("WERF_DEV_BRANCH")
	if envValue != "" {
		defaultValue = envValue
	}

	cmd.Flags().StringVarP(cmdData.DevBranch, "dev-branch", "", defaultValue, fmt.Sprintf("Set dev git branch name (default $WERF_DEV_BRANCH or %q)", defaultValue))
}

type SetupHomeDirOptions struct {
	Persistent bool
}

func SetupHomeDir(cmdData *CmdData, cmd *cobra.Command, opts SetupHomeDirOptions) {
	cmdData.HomeDir = new(string)
	getFlags(cmd, opts.Persistent).StringVarP(cmdData.HomeDir, "home-dir", "", "", "Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)")
}

func SetupSSHKey(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SSHKeys = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", []string{}, `Use only specific ssh key(s).
Can be specified with $WERF_SSH_KEY_* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa, $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa).
Defaults to $WERF_SSH_KEY_*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}`)
}

func SetupSaveBuildReport(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SaveBuildReport = new(bool)
	cmd.Flags().BoolVarP(cmdData.SaveBuildReport, "save-build-report", "", util.GetBoolEnvironmentDefaultFalse("WERF_SAVE_BUILD_REPORT"), fmt.Sprintf("Save build report (by default $WERF_SAVE_BUILD_REPORT or %t). Its path and format configured with --build-report-path", DefaultSaveBuildReport))
}

func SetupBuildReportPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.BuildReportPath = new(string)
	cmd.Flags().StringVarP(cmdData.BuildReportPath, "build-report-path", "", os.Getenv("WERF_BUILD_REPORT_PATH"), fmt.Sprintf("Change build report path and format (by default $WERF_BUILD_REPORT_PATH or %q if not set). Extension must be either .json for JSON format or .env for env-file format. If extension not specified, then .json is used", DefaultBuildReportPathJSON))
}

func SetupUseBuildReport(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.UseBuildReport = new(bool)
	cmd.Flags().BoolVarP(cmdData.UseBuildReport, "use-build-report", "", util.GetBoolEnvironmentDefaultFalse("WERF_USE_BUILD_REPORT"), fmt.Sprintf("Use build report, previously saved with --save-build-report (by default $WERF_USE_BUILD_REPORT or %t). Its path and format configured with --build-report-path", DefaultUseBuildReport))
}

func GetSaveBuildReport(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.SaveBuildReport, false)
}

func GetBuildReportPath(cmdData *CmdData) string {
	return option.PtrValueOrDefault(cmdData.BuildReportPath, "")
}

func GetUseBuildReport(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.UseBuildReport, false)
}

func GetBuildReportPathAndFormat(cmdData *CmdData) (string, build.ReportFormat, error) {
	unspecifiedPath := cmdData.BuildReportPath == nil || *cmdData.BuildReportPath == ""
	if unspecifiedPath {
		return DefaultBuildReportPathJSON, build.ReportJSON, nil
	}

	switch ext := filepath.Ext(*cmdData.BuildReportPath); ext {
	case ".json":
		return *cmdData.BuildReportPath, build.ReportJSON, nil
	case ".env":
		return *cmdData.BuildReportPath, build.ReportEnvFile, nil
	case "":
		return *cmdData.BuildReportPath + ".json", build.ReportJSON, nil
	default:
		return "", "", fmt.Errorf("invalid --build-report-path %q: extension must be either .json or .env or unspecified", *cmdData.BuildReportPath)
	}
}

func SetupSaveDeployReport(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.SaveDeployReport, "save-deploy-report", "", util.GetBoolEnvironmentDefaultFalse("WERF_SAVE_DEPLOY_REPORT"), fmt.Sprintf("Save deploy report (by default $WERF_SAVE_DEPLOY_REPORT or %t). Its path and format configured with --deploy-report-path", DefaultSaveDeployReport))
}

func SetupSaveRollbackReport(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.SaveRollbackReport, "save-rollback-report", "", util.GetBoolEnvironmentDefaultFalse("WERF_SAVE_ROLLBACK_REPORT"), fmt.Sprintf("Save rollback report (by default $WERF_SAVE_ROLLBACK_REPORT or %t). Its path and format configured with --rollback-report-path", DefaultSaveRollbackReport))
}

func SetupSaveUninstallReport(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.SaveUninstallReport, "save-uninstall-report", "", util.GetBoolEnvironmentDefaultFalse("WERF_SAVE_UNINSTALL_REPORT"), fmt.Sprintf("Save uninstall report (by default $WERF_SAVE_UNINSTALL_REPORT or %t). Its path and format configured with --uninstall-report-path", DefaultSaveUninstallReport))
}

func SetupDeployReportPath(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.DeployReportPath, "deploy-report-path", "", os.Getenv("WERF_DEPLOY_REPORT_PATH"), fmt.Sprintf("Change deploy report path and format (by default $WERF_DEPLOY_REPORT_PATH or %q if not set). Extension must be .json for JSON format. If extension not specified, then .json is used", DefaultDeployReportPathJSON))
}

func SetupRollbackReportPath(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.RollbackReportPath, "rollback-report-path", "", os.Getenv("WERF_ROLLBACK_REPORT_PATH"), fmt.Sprintf("Change rollback report path and format (by default $WERF_ROLLBACK_REPORT_PATH or %q if not set). Extension must be .json for JSON format. If extension not specified, then .json is used", DefaultRollbackReportPathJSON))
}

func SetupUninstallReportPath(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.UninstallReportPath, "uninstall-report-path", "", os.Getenv("WERF_UNINSTALL_REPORT_PATH"), fmt.Sprintf("Change uninstall report path and format (by default $WERF_UNINSTALL_REPORT_PATH or %q if not set). Extension must be .json for JSON format. If extension not specified, then .json is used", DefaultUninstallReportPathJSON))
}

func SetupUseDeployReport(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.UseDeployReport, "use-deploy-report", "", util.GetBoolEnvironmentDefaultFalse("WERF_USE_DEPLOY_REPORT"), fmt.Sprintf("Use deploy report, previously saved with --save-deploy-report (by default $WERF_USE_DEPLOY_REPORT or %t). Its path and format configured with --deploy-report-path", DefaultUseDeployReport))
}

func SetupNetworkParallelism(cmdData *CmdData, cmd *cobra.Command) {
	var defVal int
	if val, err := util.GetIntEnvVar("WERF_NETWORK_PARALLELISM"); err != nil {
		panic(fmt.Sprintf("bad WERF_NETWORK_PARALLELISM value: %s", err))
	} else if val != nil {
		defVal = int(*val)
	} else {
		defVal = common.DefaultNetworkParallelism
	}

	cmd.Flags().IntVarP(
		&cmdData.NetworkParallelism,
		"network-parallelism",
		"",
		defVal,
		fmt.Sprintf("Parallelize some network operations (default $WERF_NETWORK_PARALLELISM or %d)", common.DefaultNetworkParallelism),
	)
}

func SetupNoInstallCRDs(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.NoInstallStandaloneCRDs, "no-install-crds", "", util.GetBoolEnvironmentDefaultFalse("WERF_NO_INSTALL_CRDS"), `Do not install CRDs from "crds/" directories of installed charts (default $WERF_NO_INSTALL_CRDS)`)
}

func SetupChartProvenanceKeyring(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.ChartProvenanceKeyring, "provenance-keyring", "", os.Getenv("WERF_PROVENANCE_KEYRING"), `Path to keyring containing public keys to verify chart provenance (default $WERF_PROVENANCE_KEYRING)`)
}

func SetupChartProvenanceStrategy(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.ChartProvenanceStrategy, "provenance-strategy", "", os.Getenv("WERF_PROVENANCE_STRATEGY"), `Strategy for provenance verifying (default $WERF_PROVENANCE_STRATEGY).`)
}

func SetupExtraRuntimeAnnotations(cmdData *CmdData, cmd *cobra.Command) {
	if defVal, err := util.GetStringToStringEnvVar("WERF_RUNTIME_ANNOTATIONS"); err != nil {
		panic(fmt.Sprintf("bad WERF_RUNTIME_ANNOTATIONS value: %s", err))
	} else {
		cmd.Flags().StringToStringVarP(&cmdData.ExtraRuntimeAnnotations, "runtime-annotations", "", defVal, "Add annotations which will not trigger resource updates to all resources (default $WERF_RUNTIME_ANNOTATIONS)")
	}
}

func SetupExtraRuntimeLabels(cmdData *CmdData, cmd *cobra.Command) {
	if defVal, err := util.GetStringToStringEnvVar("WERF_RUNTIME_LABELS"); err != nil {
		panic(fmt.Sprintf("bad WERF_RUNTIME_LABELS value: %s", err))
	} else {
		cmd.Flags().StringToStringVarP(&cmdData.ExtraRuntimeLabels, "runtime-labels", "", defVal, "Add labels which will not trigger resource updates to all resources (default $WERF_RUNTIME_LABELS)")
	}
}

func SetupNoShowNotes(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.NoShowNotes, "no-notes", "", util.GetBoolEnvironmentDefaultFalse("WERF_NO_NOTES"), `Don't show release notes at the end of the release (default $WERF_NO_NOTES)`)
}

func SetupTemplatesAllowDNS(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.TemplatesAllowDNS, "templates-allow-dns", "", util.GetBoolEnvironmentDefaultFalse("WERF_TEMPLATES_ALLOW_DNS"), `Allow performing DNS requests in templating (default $WERF_TEMPLATES_ALLOW_DNS)`)
}

func SetupReleaseStorageDriver(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.ReleaseStorageDriver, "release-storage", "", util.GetFirstExistingEnvVarAsString("WERF_RELEASE_STORAGE", "HELM_DRIVER"), `How releases should be stored (default $WERF_RELEASE_STORAGE)`)
}

func SetupReleaseInfoAnnotations(cmdData *CmdData, cmd *cobra.Command) {
	if defVal, err := util.GetStringToStringEnvVar("WERF_RELEASE_INFO_ANNOTATIONS"); err != nil {
		panic(fmt.Sprintf("bad WERF_RELEASE_INFO_ANNOTATIONS value: %s", err))
	} else {
		cmd.Flags().StringToStringVarP(&cmdData.ReleaseInfoAnnotations, "release-info-annotations", "", defVal, "Add annotations to release metadata (default $WERF_RELEASE_INFO_ANNOTATIONS)")
	}
}

func SetupDefaultDeletePropagation(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.DefaultDeletePropagation, "delete-propagation", "", os.Getenv("WERF_DELETE_PROPAGATION"), fmt.Sprintf("Set default delete propagation strategy (default $WERF_DELETE_PROPAGATION or %s).", common.DefaultDeletePropagation))
}

func SetupExtraAPIVersions(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&cmdData.ExtraAPIVersions, "extra-apiversions", "", []string{}, "Extra Kubernetes API versions passed to $.Capabilities.APIVersions. Can be also set with $WERF_EXTRA_APIVERSIONS_* environment variables, values can be comma-separated")
}

func SetupNoRemoveManualChanges(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.NoRemoveManualChanges, "no-remove-manual-changes", "", util.GetBoolEnvironmentDefaultFalse("WERF_NO_REMOVE_MANUAL_CHANGES"), `Don't remove fields added manually to the resource in the cluster if fields aren't present in the manifest (default $WERF_NO_REMOVE_MANUAL_CHANGES)`)
}

func SetupReleaseLabel(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&cmdData.ReleaseLabels, "release-label", "", []string{}, `Add Helm release labels (can specify multiple). Kind of labels depends or release storage driver.
Format: labelName=labelValue.
Also, can be specified with $WERF_RELEASE_LABEL_* (e.g. $WERF_RELEASE_LABEL_1=labelName1=labelValue1, $WERF_RELEASE_LABEL_2=labelName2=labelValue2)`)
}

func SetupForceAdoption(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.ForceAdoption, "force-adoption", "", util.GetBoolEnvironmentDefaultFalse("WERF_FORCE_ADOPTION"), "Always adopt resources, even if they belong to a different Helm release (default $WERF_FORCE_ADOPTION or false)")
}

func SetupDeployGraphPath(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.InstallGraphPath, "deploy-graph-path", "", os.Getenv("WERF_DEPLOY_GRAPH_PATH"), "Save deploy graph path to the specified file (by default $WERF_DEPLOY_GRAPH_PATH). Extension must be .dot or not specified. If extension not specified, then .dot is used")
}

func SetupUninstallGraphPath(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.UninstallGraphPath, "uninstall-graph-path", "", os.Getenv("WERF_UNINSTALL_GRAPH_PATH"), "Save uninstall graph path to the specified file (by default $WERF_UNINSTALL_GRAPH_PATH). Extension must be .dot or not specified. If extension not specified, then .dot is used")
}

func SetupRollbackGraphPath(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.RollbackGraphPath, "rollback-graph-path", "", os.Getenv("WERF_ROLLBACK_GRAPH_PATH"), "Save rollback graph path to the specified file (by default $WERF_ROLLBACK_GRAPH_PATH). Extension must be .dot or not specified. If extension not specified, then .dot is used")
}

func SetupRenderSubchartNotes(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.ShowSubchartNotes, "render-subchart-notes", "", util.GetBoolEnvironmentDefaultFalse("WERF_RENDER_SUBCHART_NOTES"), "If set, render subchart notes along with the parent (by default $WERF_RENDER_SUBCHART_NOTES or false)")
}

func SetupWithoutKube(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.WithoutKube = new(bool)
	cmd.Flags().BoolVarP(cmdData.WithoutKube, "without-kube", "", util.GetBoolEnvironmentDefaultFalse("WERF_WITHOUT_KUBE"), "Do not skip deployed Kubernetes images (default $WERF_WITHOUT_KUBE)")
}

const flagNameKeepStagesBuiltWithinLastNHours = "keep-stages-built-within-last-n-hours"

func SetupKeepStagesBuiltWithinLastNHours(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.keepStagesBuiltWithinLastNHours = new(uint64)
	cmd.Flags().Uint64VarP(cmdData.keepStagesBuiltWithinLastNHours, flagNameKeepStagesBuiltWithinLastNHours, "", config.DefaultKeepImagesBuiltWithinLastNHours, fmt.Sprintf("Keep stages that were built within last hours (default $WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS or %d)", config.DefaultKeepImagesBuiltWithinLastNHours))
}

func GetKeepStagesBuiltWithinLastNHours(cmdData *CmdData, cmd *cobra.Command) *uint64 {
	envValue, err := util.GetUint64EnvVar("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS")
	if err != nil {
		panic(err)
	}

	if cmd.Flags().Changed(flagNameKeepStagesBuiltWithinLastNHours) {
		return cmdData.keepStagesBuiltWithinLastNHours
	} else if envValue != nil {
		return envValue
	} else {
		return nil
	}
}

func SetupEnvironment(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.Environment, "env", "", os.Getenv("WERF_ENV"), "Use specified environment (default $WERF_ENV)")
}

func SetupRelease(cmdData *CmdData, cmd *cobra.Command, projectConfigParsed bool) {
	var usage string
	if projectConfigParsed {
		usage = "Use specified Helm release name (default [[ project ]]-[[ env ]] template or deploy.helmRelease custom template from werf.yaml or $WERF_RELEASE)"
	} else {
		usage = "Use specified Helm release name (default $WERF_RELEASE)"
	}

	cmd.Flags().StringVarP(&cmdData.Release, "release", "", os.Getenv("WERF_RELEASE"), usage)
}

func SetupNamespace(cmdData *CmdData, cmd *cobra.Command, projectConfigParsed bool) {
	var usage string
	if projectConfigParsed {
		usage = "Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or deploy.namespace custom template from werf.yaml or $WERF_NAMESPACE)"
	} else {
		usage = "Use specified Kubernetes namespace (default $WERF_NAMESPACE)"
	}

	cmd.Flags().StringVarP(&cmdData.Namespace, "namespace", "", os.Getenv("WERF_NAMESPACE"), usage)
}

func SetupAddAnnotations(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&cmdData.ExtraAnnotations, "add-annotation", "", []string{}, `Add annotation to deploying resources (can specify multiple).
Format: annoName=annoValue.
Also, can be specified with $WERF_ADD_ANNOTATION_* (e.g. $WERF_ADD_ANNOTATION_1=annoName1=annoValue1, $WERF_ADD_ANNOTATION_2=annoName2=annoValue2)`)
}

func SetupAddLabels(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&cmdData.ExtraLabels, "add-label", "", []string{}, `Add label to deploying resources (can specify multiple).
Format: labelName=labelValue.
Also, can be specified with $WERF_ADD_LABEL_* (e.g. $WERF_ADD_LABEL_1=labelName1=labelValue1, $WERF_ADD_LABEL_2=labelName2=labelValue2)`)
}

func GetFirstExistingKubeConfigEnvVar() string {
	return util.GetFirstExistingEnvVarAsString("WERF_KUBE_CONFIG", "WERF_KUBECONFIG", "KUBECONFIG")
}

func GetFirstExistingKubeConfigBase64EnvVar() string {
	return util.GetFirstExistingEnvVarAsString("WERF_KUBE_CONFIG_BASE64", "WERF_KUBECONFIG_BASE64", "KUBECONFIG_BASE64")
}

func SetupSecondaryStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SecondaryStagesStorage = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.SecondaryStagesStorage, "secondary-repo", "", []string{}, `Specify one or multiple secondary read-only repos with images that will be used as a cache.
Also, can be specified with $WERF_SECONDARY_REPO_* (e.g. $WERF_SECONDARY_REPO_1=..., $WERF_SECONDARY_REPO_2=...)`)
}

func SetupAddCustomTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.AddCustomTag = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.AddCustomTag, "add-custom-tag", "", []string{}, `Set tag alias for the content-based tag.
The alias may contain the following shortcuts:
- %image%, %image_slug% or %image_safe_slug% to use the image name (necessary if there is more than one image in the werf config);
- %image_content_based_tag% to use a content-based tag.
For cleaning custom tags and associated content-based tag are treated as one.
Also can be defined with $WERF_ADD_CUSTOM_TAG_* (e.g. $WERF_ADD_CUSTOM_TAG_1="%image%-tag1", $WERF_ADD_CUSTOM_TAG_2="%image%-tag2")`)
}

func SetupUseCustomTag(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.UseCustomTag = new(string)
	cmd.Flags().StringVarP(cmdData.UseCustomTag, "use-custom-tag", "", os.Getenv("WERF_USE_CUSTOM_TAG"), `Use a tag alias in helm templates instead of an image content-based tag (NOT RECOMMENDED).
The alias may contain the following shortcuts:
- %image%, %image_slug% or %image_safe_slug% to use the image name (necessary if there is more than one image in the werf config);
- %image_content_based_tag% to use a content-based tag.
For cleaning custom tags and associated content-based tag are treated as one.
Also, can be defined with $WERF_USE_CUSTOM_TAG (e.g. $WERF_USE_CUSTOM_TAG="%image%-tag")`)
}

func SetupCacheStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.CacheStagesStorage = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.CacheStagesStorage, "cache-repo", "", []string{}, `Specify one or multiple cache repos with images that will be used as a cache. Cache will be populated when pushing newly built images into the primary repo and when pulling existing images from the primary repo. Cache repo will be used to pull images and to get manifests before making requests to the primary repo.
Also, can be specified with $WERF_CACHE_REPO_* (e.g. $WERF_CACHE_REPO_1=..., $WERF_CACHE_REPO_2=...)`)
}

func SetupRepoOptions(cmdData *CmdData, cmd *cobra.Command, opts RepoDataOptions) {
	SetupInsecureRegistry(cmdData, cmd)
	SetupSkipTlsVerifyRegistry(cmdData, cmd)
	SetupRepo(cmdData, cmd, opts)
}

func SetupRepo(cmdData *CmdData, cmd *cobra.Command, opts RepoDataOptions) {
	cmdData.Repo = NewRepoData("repo", opts)
	cmdData.Repo.SetupCmd(cmd)
}

func SetupFinalRepo(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.FinalRepo = NewRepoData("final-repo", RepoDataOptions{})
	cmdData.FinalRepo.SetupCmd(cmd)
}

func SetupReleasesHistoryMax(cmdData *CmdData, cmd *cobra.Command) {
	defaultValueP, err := util.GetIntEnvVar("WERF_RELEASES_HISTORY_MAX")
	if err != nil {
		panic(fmt.Sprintf("bad WERF_RELEASES_HISTORY_MAX value: %s", err))
	}

	var defaultValue int
	if defaultValueP != nil {
		defaultValue = int(*defaultValueP)
	} else {
		defaultValue = 5
	}

	cmd.Flags().IntVarP(
		&cmdData.ReleaseHistoryLimit,
		"releases-history-max",
		"",
		defaultValue,
		"Max releases to keep in release storage ($WERF_RELEASES_HISTORY_MAX or 5 by default)",
	)
}

func SetupInsecureRegistry(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.InsecureRegistry != nil {
		return
	}

	cmdData.InsecureRegistry = new(bool)
	cmd.Flags().BoolVarP(cmdData.InsecureRegistry, "insecure-registry", "", util.GetBoolEnvironmentDefaultFalse("WERF_INSECURE_REGISTRY"), "Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)")
}

func SetupContainerRegistryMirror(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ContainerRegistryMirror = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.ContainerRegistryMirror, "container-registry-mirror", "", []string{}, "(Buildah-only) Use specified mirrors for docker.io")
}

func SetupSkipTlsVerifyRegistry(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.SkipTlsVerifyRegistry != nil {
		return
	}

	cmdData.SkipTlsVerifyRegistry = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipTlsVerifyRegistry, "skip-tls-verify-registry", "", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_TLS_VERIFY_REGISTRY"), "Skip TLS certificate validation when accessing a registry (default $WERF_SKIP_TLS_VERIFY_REGISTRY)")
}

func SetupReleaseStorageSQLConnection(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.ReleaseStorageSQLConnection, "release-storage-sql-connection", "", os.Getenv("WERF_RELEASE_STORAGE_SQL_CONNECTION"), "SQL Connection String for Helm SQL Storage (default $WERF_RELEASE_STORAGE_SQL_CONNECTION)")
}

func SetupDryRun(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DryRun = new(bool)
	cmd.Flags().BoolVarP(cmdData.DryRun, "dry-run", "", util.GetBoolEnvironmentDefaultFalse("WERF_DRY_RUN"), "Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)")
}

func SetupDockerConfig(cmdData *CmdData, cmd *cobra.Command, extraDesc string) {
	defaultValue := os.Getenv("WERF_DOCKER_CONFIG")
	if defaultValue == "" {
		defaultValue = os.Getenv("DOCKER_CONFIG")
	}

	cmdData.DockerConfig = new(string)

	desc := "Specify docker config directory path. Default $WERF_DOCKER_CONFIG or $DOCKER_CONFIG or ~/.docker (in the order of priority)"

	if extraDesc != "" {
		desc += "\n"
		desc += extraDesc
	}

	cmd.Flags().StringVarP(cmdData.DockerConfig, "docker-config", "", defaultValue, desc)
}

func SetupLogOptions(cmdData *CmdData, cmd *cobra.Command) {
	setupLogOptions(cmdData, cmd, false)
}

func SetupLogOptionsDefaultQuiet(cmdData *CmdData, cmd *cobra.Command) {
	setupLogOptions(cmdData, cmd, true)
}

func setupLogOptions(cmdData *CmdData, cmd *cobra.Command, defaultQuiet bool) {
	setupLogDebug(cmdData, cmd)
	setupLogVerbose(cmdData, cmd)
	setupLogQuiet(cmdData, cmd, defaultQuiet)
	setupLogColor(cmdData, cmd)
	setupLogPretty(cmdData, cmd)
	setupLogTime(cmdData, cmd)
	setupLogTimeFormat(cmdData, cmd)
	setupTerminalWidth(cmdData, cmd)
}

func setupLogDebug(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogDebug = new(bool)

	defaultValue := false
	for _, envName := range []string{
		"WERF_LOG_DEBUG",
		"WERF_DEBUG",
	} {
		if os.Getenv(envName) != "" {
			defaultValue = util.GetBoolEnvironmentDefaultFalse(envName)
			break
		}
	}

	for alias, env := range map[string]string{
		"log-debug": "WERF_LOG_DEBUG",
		"debug":     "WERF_DEBUG",
	} {
		cmd.PersistentFlags().BoolVarP(
			cmdData.LogDebug,
			alias,
			"",
			defaultValue,
			fmt.Sprintf("Enable debug (default $%s).", env),
		)
	}

	if err := cmd.PersistentFlags().MarkHidden("debug"); err != nil {
		panic(err)
	}
}

func setupLogColor(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogColorMode = new(string)

	logColorEnvironmentValue := os.Getenv("WERF_LOG_COLOR_MODE")

	defaultValue := "auto"
	if logColorEnvironmentValue != "" {
		defaultValue = logColorEnvironmentValue
	}

	cmd.PersistentFlags().StringVarP(cmdData.LogColorMode, "log-color-mode", "", defaultValue, `Set log color mode.
Supported on, off and auto (based on the stdout’s file descriptor referring to a terminal) modes.
Default $WERF_LOG_COLOR_MODE or auto mode.`)
}

func setupLogQuiet(cmdData *CmdData, cmd *cobra.Command, isDefaultQuiet bool) {
	cmdData.LogQuiet = new(bool)

	defaultValue := isDefaultQuiet

	for _, envName := range []string{
		"WERF_LOG_QUIET",
		"WERF_QUIET",
	} {
		if os.Getenv(envName) != "" {
			if defaultValue {
				defaultValue = util.GetBoolEnvironmentDefaultTrue(envName)
			} else {
				defaultValue = util.GetBoolEnvironmentDefaultFalse(envName)
			}

			break
		}
	}

	for alias, env := range map[string]string{
		"log-quiet": "WERF_LOG_QUIET",
		"quiet":     "WERF_QUIET",
	} {
		cmd.PersistentFlags().BoolVarP(
			cmdData.LogQuiet,
			alias,
			"",
			defaultValue,
			fmt.Sprintf(`Disable explanatory output (default $%s).`, env),
		)
	}

	if err := cmd.PersistentFlags().MarkHidden("quiet"); err != nil {
		panic(err)
	}
}

func setupLogVerbose(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogVerbose = new(bool)

	var defaultValue bool
	for _, envName := range []string{
		"WERF_LOG_VERBOSE",
		"WERF_VERBOSE",
	} {
		if os.Getenv(envName) != "" {
			defaultValue = util.GetBoolEnvironmentDefaultFalse(envName)
			break
		}
	}

	for alias, env := range map[string]string{
		"log-verbose": "WERF_LOG_VERBOSE",
		"verbose":     "WERF_VERBOSE",
	} {
		cmd.PersistentFlags().BoolVarP(
			cmdData.LogVerbose,
			alias,
			"",
			defaultValue,
			fmt.Sprintf(`Enable verbose output (default $%s).`, env),
		)
	}

	if err := cmd.PersistentFlags().MarkHidden("verbose"); err != nil {
		panic(err)
	}
}

func setupLogPretty(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogPretty = new(bool)

	var defaultValue bool
	if os.Getenv("WERF_LOG_PRETTY") != "" {
		defaultValue = util.GetBoolEnvironmentDefaultFalse("WERF_LOG_PRETTY")
	} else {
		defaultValue = true
	}

	cmd.PersistentFlags().BoolVarP(cmdData.LogPretty, "log-pretty", "", defaultValue, `Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or true).`)
}

func setupLogTime(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogTime = new(bool)
	cmd.PersistentFlags().BoolVarP(cmdData.LogTime, "log-time", "", util.GetBoolEnvironmentDefaultFalse("WERF_LOG_TIME"), `Add time to log entries for precise event time tracking (default $WERF_LOG_TIME or false).`)
}

func setupLogTimeFormat(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogTimeFormat = new(string)

	defaultValue := os.Getenv("WERF_LOG_TIME_FORMAT")
	if defaultValue == "" {
		defaultValue = time.RFC3339
	}
	cmd.PersistentFlags().StringVarP(cmdData.LogTimeFormat, "log-time-format", "", defaultValue, `Specify custom log time format (default $WERF_LOG_TIME_FORMAT or RFC3339 format).`)
}

func setupTerminalWidth(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogTerminalWidth = new(int64)
	cmd.PersistentFlags().Int64VarP(cmdData.LogTerminalWidth, "log-terminal-width", "", -1, fmt.Sprintf(`Set log terminal width.
Defaults to:
* $WERF_LOG_TERMINAL_WIDTH
* interactive terminal width or %d`, 140))
}

func SetupLogProjectDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogProjectDir = new(bool)
	cmd.Flags().BoolVarP(cmdData.LogProjectDir, "log-project-dir", "", util.GetBoolEnvironmentDefaultFalse("WERF_LOG_PROJECT_DIR"), `Print current project directory path (default $WERF_LOG_PROJECT_DIR)`)
}

func SetupIntrospectAfterError(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IntrospectAfterError = new(bool)
	cmd.Flags().BoolVarP(cmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
}

func GetIntrospectAfterError(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.IntrospectAfterError, false)
}

func SetupIntrospectBeforeError(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IntrospectBeforeError = new(bool)
	cmd.Flags().BoolVarP(cmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")
}

func GetIntrospectBeforeError(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.IntrospectBeforeError, false)
}

func GetIntrospectStage(cmdData *CmdData) []string {
	return option.PtrValueOrDefault(cmdData.StagesToIntrospect, []string{})
}

func SetupIntrospectStage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesToIntrospect = new([]string)
	cmd.Flags().StringArrayVarP(cmdData.StagesToIntrospect, "introspect-stage", "", []string{}, `Introspect a specific stage. The option can be used multiple times to introspect several stages.

There are the following formats to use:
* specify IMAGE_NAME/STAGE_NAME to introspect stage STAGE_NAME of either image or artifact IMAGE_NAME
* specify STAGE_NAME or */STAGE_NAME for the introspection of all existing stages with name STAGE_NAME

IMAGE_NAME is the name of an image or artifact described in werf.yaml, the nameless image specified with ~.
STAGE_NAME should be one of the following: `+strings.Join(allStagesNames(), ", "))
}

// SetupRequireBuiltImages adds --require-built-images flag.
// See also [quireBuiltImages].
func SetupRequireBuiltImages(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.RequireBuiltImages = new(bool)
	cmd.Flags().BoolVarP(cmdData.RequireBuiltImages, "require-built-images", "Z", util.GetBoolEnvironmentDefaultFalse("WERF_REQUIRE_BUILT_IMAGES"), "Requires all used images to be previously built and exist in repo. Exits with error if needed images are not cached and so require to run build instructions (default $WERF_REQUIRE_BUILT_IMAGES)")
}

func SetupCheckBuiltImages(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.CheckBuiltImages = new(bool)
	cmd.Flags().BoolVarP(cmdData.CheckBuiltImages, "check-built-images", "", util.GetBoolEnvironmentDefaultFalse("WERF_CHECK_BUILT_IMAGES"), "Check that all used images are previously built and exist in repo. Exits with error if needed images are not cached and so require to run build instructions (default $WERF_CHECK_BUILT_IMAGES)")

	cmdData.LegacyCheckBuiltImages = new(bool)
	cmd.Flags().BoolVarP(cmdData.LegacyCheckBuiltImages, "require-built-images", "Z", false, "Check that all used images are previously built and exist in repo. Exits with error if needed images are not cached and so require to run build instructions")
	if err := cmd.Flags().MarkHidden("require-built-images"); err != nil {
		panic(err)
	}
}

func SetupStubTags(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StubTags = new(bool)
	cmd.Flags().BoolVarP(cmdData.StubTags, "stub-tags", "", util.GetBoolEnvironmentDefaultFalse("WERF_STUB_TAGS"), "Use stubs instead of real tags (default $WERF_STUB_TAGS)")
}

func SetupFollow(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Follow = new(bool)
	cmd.Flags().BoolVarP(cmdData.Follow, "follow", "", util.GetBoolEnvironmentDefaultFalse("WERF_FOLLOW"), `Enable follow mode (default $WERF_FOLLOW).
The mode allows restarting the command on a new commit.
In development mode (--dev), werf restarts the command on any changes (including untracked files) in the git repository worktree`)
}

func SetupKubeVersion(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cmdData.KubeVersion, "kube-version", "", os.Getenv("WERF_KUBE_VERSION"), "Set specific Capabilities.KubeVersion (default $WERF_KUBE_VERSION)")
}

func allStagesNames() []string {
	var stageNames []string
	for _, stageName := range stage.AllStages {
		stageNames = append(stageNames, string(stageName))
	}

	return stageNames
}

func GetLocalStagesStorage(containerBackend container_backend.ContainerBackend) *storage.LocalStagesStorage {
	return storage.NewLocalStagesStorage(containerBackend)
}

type GetStagesStorageOpts struct {
	CleanupDisabled                bool
	GitHistoryBasedCleanupDisabled bool
	SkipMetaCheck                  bool
}

func GetStagesStorage(ctx context.Context, containerBackend container_backend.ContainerBackend, cmdData *CmdData, opts GetStagesStorageOpts) (storage.PrimaryStagesStorage, error) {
	return cmdData.Repo.CreateStagesStorage(ctx, &CreateStagesStorageOptions{
		ContainerBackend:               containerBackend,
		InsecureRegistry:               *cmdData.InsecureRegistry,
		SkipTlsVerifyRegistry:          *cmdData.SkipTlsVerifyRegistry,
		CleanupDisabled:                opts.CleanupDisabled,
		GitHistoryBasedCleanupDisabled: opts.GitHistoryBasedCleanupDisabled,
		SkipMetaCheck:                  opts.SkipMetaCheck,
	})
}

func GetOptionalFinalStagesStorage(ctx context.Context, containerBackend container_backend.ContainerBackend, cmdData *CmdData) (storage.StagesStorage, error) {
	if *cmdData.FinalRepo.Address == "" {
		return nil, nil
	}
	return cmdData.FinalRepo.CreateStagesStorage(ctx, &CreateStagesStorageOptions{
		ContainerBackend:      containerBackend,
		InsecureRegistry:      *cmdData.InsecureRegistry,
		SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
		SkipMetaCheck:         true,
	})
}

func GetCacheStagesStorageList(ctx context.Context, containerBackend container_backend.ContainerBackend, cmdData *CmdData) ([]storage.StagesStorage, error) {
	var res []storage.StagesStorage

	for _, address := range GetCacheStagesStorage(cmdData) {
		repoData := NewRepoData("cache-repo", RepoDataOptions{OnlyAddress: true})
		repoData.Address = &address

		storage, err := repoData.CreateStagesStorage(ctx, &CreateStagesStorageOptions{
			ContainerBackend:      containerBackend,
			InsecureRegistry:      *cmdData.InsecureRegistry,
			SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
			SkipMetaCheck:         true,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create cache stages storage in %s: %w", address, err)
		}
		res = append(res, storage)
	}

	return res, nil
}

func GetSecondaryStagesStorageList(ctx context.Context, stagesStorage storage.StagesStorage, containerBackend container_backend.ContainerBackend, cmdData *CmdData) ([]storage.StagesStorage, error) {
	var res []storage.StagesStorage

	if stagesStorage.Address() != storage.LocalStorageAddress {
		res = append(res, storage.NewLocalStagesStorage(containerBackend))
	}

	for _, address := range GetSecondaryStagesStorage(cmdData) {
		repoData := NewRepoData("secondary-repo", RepoDataOptions{OnlyAddress: true})
		repoData.Address = &address

		storage, err := repoData.CreateStagesStorage(ctx, &CreateStagesStorageOptions{
			ContainerBackend:      containerBackend,
			InsecureRegistry:      *cmdData.InsecureRegistry,
			SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
			SkipMetaCheck:         true,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create secondary stages storage in %s: %w", address, err)
		}
		res = append(res, storage)
	}

	return res, nil
}

func GetOptionalWerfConfig(ctx context.Context, cmdData *CmdData, giterminismManager giterminism_manager.Interface, opts config.WerfConfigOptions) (string, *config.WerfConfig, error) {
	customWerfConfigRelPath, err := GetCustomWerfConfigRelPath(giterminismManager, cmdData)
	if err != nil {
		return "", nil, err
	}

	exist, err := giterminismManager.FileReader().IsConfigExistAnywhere(ctx, customWerfConfigRelPath)
	if err != nil {
		return "", nil, err
	}

	if exist {
		customWerfConfigTemplatesDirRelPath, err := GetCustomWerfConfigTemplatesDirRelPath(giterminismManager, cmdData)
		if err != nil {
			return "", nil, err
		}

		customWerfConfigRenderPath, err := GetCustomWerfConfigRenderPath(cmdData)
		if err != nil {
			return "", nil, err
		}

		configPath, c, err := config.GetWerfConfig(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, customWerfConfigRenderPath, giterminismManager, opts)
		if err != nil {
			return "", nil, err
		}

		return configPath, c, nil
	}

	return "", nil, nil
}

func GetRequiredWerfConfig(ctx context.Context, cmdData *CmdData, giterminismManager giterminism_manager.Interface, opts config.WerfConfigOptions) (string, *config.WerfConfig, error) {
	customWerfConfigRelPath, err := GetCustomWerfConfigRelPath(giterminismManager, cmdData)
	if err != nil {
		return "", nil, err
	}

	customWerfConfigTemplatesDirRelPath, err := GetCustomWerfConfigTemplatesDirRelPath(giterminismManager, cmdData)
	if err != nil {
		return "", nil, err
	}

	customWerfConfigRenderPath, err := GetCustomWerfConfigRenderPath(cmdData)
	if err != nil {
		return "", nil, err
	}

	configPath, c, err := config.GetWerfConfig(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, customWerfConfigRenderPath, giterminismManager, opts)
	if err != nil {
		return "", nil, err
	}

	return configPath, c, nil
}

func GetCustomWerfConfigRelPath(giterminismManager giterminism_manager.Interface, cmdData *CmdData) (string, error) {
	customConfigPath := *cmdData.ConfigPath
	if customConfigPath == "" {
		return "", nil
	}

	customConfigPath = util.GetAbsoluteFilepath(customConfigPath)
	if !util.IsSubpathOfBasePath(giterminismManager.LocalGitRepo().GetWorkTreeDir(), customConfigPath) {
		return "", fmt.Errorf("the werf config %q must be in the project git work tree %q", customConfigPath, giterminismManager.LocalGitRepo().GetWorkTreeDir())
	}

	return util.GetRelativeToBaseFilepath(giterminismManager.ProjectDir(), customConfigPath), nil
}

func GetWerfGiterminismConfigRelPath(cmdData *CmdData) string {
	path := cmdData.GiterminismConfigRelPath
	if path == nil || *path == "" {
		return "werf-giterminism.yaml"
	}

	return filepath.ToSlash(*path)
}

func GetCustomWerfConfigTemplatesDirRelPath(giterminismManager giterminism_manager.Interface, cmdData *CmdData) (string, error) {
	customConfigTemplatesDirPath := *cmdData.ConfigTemplatesDir
	if customConfigTemplatesDirPath == "" {
		return "", nil
	}

	customConfigTemplatesDirPath = util.GetAbsoluteFilepath(customConfigTemplatesDirPath)
	if !util.IsSubpathOfBasePath(giterminismManager.LocalGitRepo().GetWorkTreeDir(), customConfigTemplatesDirPath) {
		return "", fmt.Errorf("the werf configuration templates directory %q must be in the project git work tree %q", customConfigTemplatesDirPath, giterminismManager.LocalGitRepo().GetWorkTreeDir())
	}

	return util.GetRelativeToBaseFilepath(giterminismManager.ProjectDir(), customConfigTemplatesDirPath), nil
}

func GetCustomWerfConfigRenderPath(cmdData *CmdData) (string, error) {
	if cmdData.ConfigRenderPath == nil || *cmdData.ConfigRenderPath == "" {
		return "", nil
	}

	customConfigRenderPath := *cmdData.ConfigRenderPath
	customConfigRenderPath = util.GetAbsoluteFilepath(customConfigRenderPath)

	return customConfigRenderPath, nil
}

func GetWerfConfigOptions(cmdData *CmdData, logRenderedFilePath bool) config.WerfConfigOptions {
	return config.WerfConfigOptions{
		LogRenderedFilePath: logRenderedFilePath,
		Env:                 cmdData.Environment,
		DebugTemplates:      cmdData.DebugTemplates,
	}
}

func OpenGitRepo(ctx context.Context, cmdData *CmdData, workingDir, gitWorkTree string) (*git_repo.Local, error) {
	isWorkingDirInsideGitWorkTree := util.IsSubpathOfBasePath(gitWorkTree, workingDir)
	areWorkingDirAndGitWorkTreeTheSame := gitWorkTree == workingDir
	if !(isWorkingDirInsideGitWorkTree || areWorkingDirAndGitWorkTreeTheSame) {
		return nil, fmt.Errorf("werf requires project dir — the current working directory or directory specified with --dir option (or WERF_DIR env var) — to be located inside the git work tree: %q is located outside of the git work tree %q", gitWorkTree, workingDir)
	}

	var openLocalRepoOptions git_repo.OpenLocalRepoOptions
	if *cmdData.Dev {
		openLocalRepoOptions.WithServiceHeadCommit = true
		openLocalRepoOptions.ServiceBranchOptions.Name = *cmdData.DevBranch
		openLocalRepoOptions.ServiceBranchOptions.GlobExcludeList = GetDevIgnore(cmdData)
	}

	return git_repo.OpenLocalRepo(ctx, "own", gitWorkTree, openLocalRepoOptions)
}

func GetGiterminismManager(ctx context.Context, cmdData *CmdData) (*giterminism_manager.Manager, error) {
	printGlobalWarningIfDevInCI(ctx, cmdData)
	manager := new(giterminism_manager.Manager)
	if err := logboek.Context(ctx).Info().LogProcess("Initialize giterminism manager").
		DoError(func() error {
			workingDir := GetWorkingDir(cmdData)

			gitWorkTree, err := GetGitWorkTree(ctx, cmdData, workingDir)
			if err != nil {
				return fmt.Errorf("unable to get git work tree: %w", err)
			}

			localGitRepo, err := OpenGitRepo(ctx, cmdData, workingDir, gitWorkTree)
			if err != nil {
				return err
			}

			headCommit, err := localGitRepo.HeadCommitHash(ctx)
			if err != nil {
				return err
			}

			configRelPath := GetWerfGiterminismConfigRelPath(cmdData)

			gm, err := giterminism_manager.NewManager(ctx, configRelPath, workingDir, localGitRepo, headCommit, giterminism_manager.NewManagerOptions{
				LooseGiterminism:       *cmdData.LooseGiterminism,
				Dev:                    *cmdData.Dev,
				CreateIncludesLockFile: cmdData.CreateIncludesLockFile,
				AllowIncludesUpdate:    cmdData.AllowIncludesUpdate,
			})
			if err != nil {
				return err
			}

			manager = gm

			return nil
		}); err != nil {
		return nil, err
	}

	return manager, nil
}

func GetGitWorkTree(ctx context.Context, cmdData *CmdData, workingDir string) (string, error) {
	if *cmdData.GitWorkTree != "" {
		workTree := *cmdData.GitWorkTree

		if isValid, err := true_git.IsValidWorkTree(ctx, workTree); err != nil {
			return "", err
		} else if isValid {
			return util.GetAbsoluteFilepath(workTree), nil
		}

		return "", fmt.Errorf("werf requires a git work tree for the project to exist: not a valid git work tree %q specified", workTree)
	}

	if found, workTree, err := true_git.UpwardLookupAndVerifyWorkTree(ctx, workingDir); err != nil {
		return "", err
	} else if found {
		return util.GetAbsoluteFilepath(workTree), nil
	}

	if res, err := LookupGitWorkTree(ctx, workingDir); err != nil {
		return "", fmt.Errorf("unable to lookup git work tree from wd %q: %w", workingDir, err)
	} else if res != "" {
		return res, nil
	}

	return "", &GitWorktreeNotFoundError{}
}

func printGlobalWarningIfDevInCI(ctx context.Context, cmdData *CmdData) {
	const (
		devModeInCIWarning = `The development mode is enabled in CI environment by providing --dev flag or WERF_DEV env variable.
This mode is intended for local development only and relies on the Git worktree state, including tracked and untracked files, while ignoring changes based on .gitignore and --dev-ignore rules.
Using development in CI may lead to non-reproducible builds and unexpected results.`
	)
	if cmdData.Dev != nil && *cmdData.Dev {
		if werf.IsRunningInCI() {
			global_warnings.GlobalWarningLn(ctx, devModeInCIWarning)
		}
	}
}

func LookupGitWorkTree(ctx context.Context, workingDir string) (string, error) {
	if found, workTree, err := true_git.UpwardLookupAndVerifyWorkTree(ctx, workingDir); err != nil {
		return "", err
	} else if found {
		return util.GetAbsoluteFilepath(workTree), nil
	}

	return "", nil
}

func GetWorkingDir(cmdData *CmdData) string {
	var workingDir string
	if *cmdData.Dir != "" {
		workingDir = *cmdData.Dir
	} else {
		workingDir = "."
	}
	return util.GetAbsoluteFilepath(workingDir)
}

func GetHelmChartDir(werfConfigPath string, werfConfig *config.WerfConfig, giterminismManager giterminism_manager.Interface) (string, error) {
	var helmChartDir string
	if werfConfig.Meta.Deploy.HelmChartDir != nil && *werfConfig.Meta.Deploy.HelmChartDir != "" {
		helmChartDir = *werfConfig.Meta.Deploy.HelmChartDir
	} else {
		helmChartDir = filepath.Join(filepath.Dir(werfConfigPath), ".helm")
	}

	absHelmChartDir := filepath.Join(giterminismManager.ProjectDir(), helmChartDir)
	if !util.IsSubpathOfBasePath(giterminismManager.LocalGitRepo().GetWorkTreeDir(), absHelmChartDir) {
		return "", fmt.Errorf("the chart directory %s must be in the project git work tree %s", absHelmChartDir, giterminismManager.LocalGitRepo().GetWorkTreeDir())
	}

	return helmChartDir, nil
}

func GetHelmChartConfigAppVersion(werfConfig *config.WerfConfig) string {
	if werfConfig.Meta.Deploy.HelmChartConfig.AppVersion != nil {
		return *werfConfig.Meta.Deploy.HelmChartConfig.AppVersion
	}

	return ""
}

func GetNamespace(cmdData *CmdData) string {
	if cmdData.Namespace == "" {
		return "default"
	}
	return cmdData.Namespace
}

func GetDevIgnore(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_DEV_IGNORE_"), *cmdData.DevIgnore...)
}

func GetSSHKey(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_SSH_KEY_"), *cmdData.SSHKeys...)
}

func GetAddLabels(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_ADD_LABEL_"), cmdData.ExtraLabels...)
}

func GetAddAnnotations(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_ADD_ANNOTATION_"), cmdData.ExtraAnnotations...)
}

func GetReleaseLabel(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_RELEASE_LABEL_"), cmdData.ReleaseLabels...)
}

func GetCacheStagesStorage(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_CACHE_REPO_"), *cmdData.CacheStagesStorage...)
}

func GetSecondaryStagesStorage(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_SECONDARY_REPO_"), *cmdData.SecondaryStagesStorage...)
}

func GetContainerRegistryMirror(ctx context.Context, cmdData *CmdData) ([]string, error) {
	cmdMirrors := append(util.PredefinedValuesByEnvNamePrefix("WERF_CONTAINER_REGISTRY_MIRROR_"), *cmdData.ContainerRegistryMirror...)

	dockerMirrors, err := docker.GetRegistryMirrors(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get docker registry mirrors: %w", err)
	}

	var result []string
	seen := make(map[string]bool)

	for _, mirror := range cmdMirrors {
		if strings.HasPrefix(mirror, "http://") {
			return nil, fmt.Errorf("invalid container registry mirror %q: only https schema allowed", mirror)
		}

		if !strings.HasPrefix(mirror, "http://") && !strings.HasPrefix(mirror, "https://") {
			mirror = "https://" + mirror
		}

		if !seen[mirror] {
			seen[mirror] = true
			result = append(result, mirror)
		}
	}

	for _, mirror := range dockerMirrors {
		if !seen[mirror] {
			seen[mirror] = true
			result = append(result, mirror)
		}
	}

	return result, nil
}

func getAddCustomTag(cmdData *CmdData) []string {
	return append(util.PredefinedValuesByEnvNamePrefix("WERF_ADD_CUSTOM_TAG_"), *cmdData.AddCustomTag...)
}

func GetRequiredRelease(cmdData *CmdData) (string, error) {
	if cmdData.Release == "" {
		return "", fmt.Errorf("--release=RELEASE param required")
	}
	return cmdData.Release, nil
}

func GetOptionalRelease(cmdData *CmdData) string {
	if cmdData.Release == "" {
		return "werf-stub"
	}
	return cmdData.Release
}

// GetRequireBuiltImages returns true if --require-built-images is set or --skip-build is set.
// There is no way to determine if both options are used, so no warning.
func GetRequireBuiltImages(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.RequireBuiltImages, false)
}

func GetIntrospectOptions(cmdData *CmdData, werfConfig *config.WerfConfig) (build.IntrospectOptions, error) {
	isStageExist := func(sName string) bool {
		for _, stageName := range allStagesNames() {
			if sName == stageName {
				return true
			}
		}

		return false
	}

	introspectOptions := build.IntrospectOptions{}

	for _, optionValue := range GetIntrospectStage(cmdData) {
		var imageName, stageName string
		{
			parts := strings.SplitN(optionValue, "/", 2)
			if len(parts) == 1 {
				imageName = "*"
				stageName = parts[0]
			} else {
				if parts[0] != "~" {
					imageName = parts[0]
				}

				stageName = parts[1]
			}
		}

		if imageName != "*" && werfConfig.GetImage(imageName) == nil {
			return introspectOptions, fmt.Errorf("specified image %q (%q) is not defined in werf.yaml", imageName, optionValue)
		}

		if !isStageExist(stageName) {
			return introspectOptions, fmt.Errorf("specified stage name %q (%q) is not exist", stageName, optionValue)
		}

		introspectTarget := build.IntrospectTarget{ImageName: imageName, StageName: stageName}
		introspectOptions.Targets = append(introspectOptions.Targets, introspectTarget)
	}

	return introspectOptions, nil
}

func LogKubeContext(kubeContext string) {
	if kubeContext != "" {
		logboek.LogF("Using kube context: %s\n", kubeContext)
	}
}

func ProcessLogProjectDir(cmdData *CmdData, projectDir string) {
	if *cmdData.LogProjectDir {
		logboek.LogF("Using project dir: %s\n", projectDir)
	}
}

func ProcessLogOptions(cmdData *CmdData) error {
	if err := ProcessLogColorMode(cmdData); err != nil {
		return err
	}

	switch {
	case *cmdData.LogDebug:
		logboek.SetAcceptedLevel(level.Debug)
		logboek.Streams().EnablePrefixDuration()
		logboek.Streams().SetPrefixStyle(style.Details())
	case *cmdData.LogVerbose:
		logboek.SetAcceptedLevel(level.Info)
	case *cmdData.LogQuiet:
		logboek.SetAcceptedLevel(level.Error)
	}

	if *cmdData.LogTime {
		logboek.Streams().EnablePrefixTime()
		logboek.Streams().SetPrefixTimeFormat(*cmdData.LogTimeFormat)
		logboek.Streams().SetPrefixStyle(style.Details())
	}

	if !*cmdData.LogPretty {
		logboek.Streams().DisablePrettyLog()
		logging.DisablePrettyLog()
	}

	if err := ProcessLogTerminalWidth(cmdData); err != nil {
		return err
	}

	return nil
}

func GetNelmLogLevel(cmdData *CmdData) log.Level {
	if util.GetBoolEnvironmentDefaultFalse("WERF_NELM_TRACE") {
		return log.TraceLevel
	}

	var logLevel log.Level
	switch {
	case *cmdData.LogDebug:
		logLevel = log.DebugLevel
	case *cmdData.LogQuiet:
		logLevel = log.ErrorLevel
	default:
		logLevel = log.InfoLevel
	}

	return logLevel
}

func ProcessLogColorMode(cmdData *CmdData) error {
	logColorMode := *cmdData.LogColorMode

	switch logColorMode {
	case "auto":
	case "on":
		logboek.Streams().EnableStyle()
	case "off":
		logboek.Streams().DisableStyle()
	default:
		return fmt.Errorf("bad log color mode %q: on, off and auto modes are supported", logColorMode)
	}

	return nil
}

func ProcessLogTerminalWidth(cmdData *CmdData) error {
	value := *cmdData.LogTerminalWidth

	if value != -1 {
		if value < 0 {
			return fmt.Errorf("--log-terminal-width parameter (%d) can not be negative", value)
		}

		logboek.Streams().SetWidth(int(value))
	} else {
		pInt64, err := util.GetInt64EnvVar("WERF_LOG_TERMINAL_WIDTH")
		if err != nil {
			return err
		}

		if pInt64 == nil {
			return nil
		}

		if *pInt64 < 0 {
			return fmt.Errorf("WERF_LOG_TERMINAL_WIDTH value (%s) can not be negative", os.Getenv("WERF_LOG_TERMINAL_WIDTH"))
		}

		logboek.Streams().SetWidth(int(*pInt64))
	}

	return nil
}

func DockerRegistryInit(ctx context.Context, cmdData *CmdData, registryMirrors []string) error {
	return docker_registry.Init(ctx, *cmdData.InsecureRegistry, *cmdData.SkipTlsVerifyRegistry, registryMirrors)
}

func ValidateMinimumNArgs(minArgs int, args []string, cmd *cobra.Command) error {
	if len(args) < minArgs {
		PrintHelp(cmd)
		return fmt.Errorf("requires at least %d arg(s), received %d", minArgs, len(args))
	}

	return nil
}

func ValidateArgumentCount(expectedCount int, args []string, cmd *cobra.Command) error {
	if len(args) != expectedCount {
		PrintHelp(cmd)
		return fmt.Errorf("requires %d position argument(s), received %d", expectedCount, len(args))
	}

	return nil
}

func PrintHelp(cmd *cobra.Command) {
	_ = cmd.Help()
	logboek.LogOptionalLn()
}

func LogRunningTime(f func() error) error {
	t := time.Now()
	err := f()

	logboek.Default().LogFHighlight("Running time %0.2f seconds\n", time.Since(t).Seconds())

	return err
}

func LogVersion() {
	logboek.LogF("Version: %s\n", werf.Version)
}

func SetupVirtualMerge(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMerge = new(bool)
	cmd.Flags().BoolVarP(cmdData.VirtualMerge, "virtual-merge", "", util.GetBoolEnvironmentDefaultFalse("WERF_VIRTUAL_MERGE"), "Enable virtual/ephemeral merge commit mode when building current application state ($WERF_VIRTUAL_MERGE by default)")
}

func GetVirtualMerge(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.VirtualMerge, false)
}

func getFlags(cmd *cobra.Command, persistent bool) *pflag.FlagSet {
	if persistent {
		return cmd.PersistentFlags()
	}

	return cmd.Flags()
}

// TODO(major): get rid of this, don't require Kubernetes for non-deployment related tasks
func SetupMinimalKubeConnectionFlags(cmdData *CmdData, cmd *cobra.Command) error {
	SetupKubeConfigBase64(cmdData, cmd)
	SetupLegacyKubeConfigPath(cmdData, cmd)
	SetupKubeContextCurrent(cmdData, cmd)

	return nil
}

func SetupKubeConnectionFlags(cmdData *CmdData, cmd *cobra.Command) error {
	cmd.Flags().StringVarP(&cmdData.KubeAPIServerAddress, "kube-api-server", "", os.Getenv("WERF_KUBE_API_SERVER"), "Kubernetes API server address (default $WERF_KUBE_API_SERVER)")
	if defVal, err := util.GetStringToStringEnvVar("WERF_KUBE_AUTH_PROVIDER_CONFIG"); err != nil {
		return fmt.Errorf("bad WERF_KUBE_AUTH_PROVIDER_CONFIG value: %w", err)
	} else {
		cmd.Flags().StringToStringVarP(&cmdData.KubeAuthProviderConfig, "kube-auth-provider-config", "", defVal, "Auth provider config for authentication in Kubernetes API (default $WERF_KUBE_AUTH_PROVIDER_CONFIG)")
	}
	cmd.Flags().StringVarP(&cmdData.KubeAuthProviderName, "kube-auth-provider", "", os.Getenv("WERF_KUBE_AUTH_PROVIDER"), "Auth provider name for authentication in Kubernetes API (default $WERF_KUBE_AUTH_PROVIDER)")
	cmd.Flags().StringVarP(&cmdData.KubeBasicAuthPassword, "kube-auth-password", "", os.Getenv("WERF_KUBE_AUTH_PASSWORD"), "Basic auth password for Kubernetes API (default $WERF_KUBE_AUTH_PASSWORD)")
	cmd.Flags().StringVarP(&cmdData.KubeBasicAuthUsername, "kube-auth-username", "", os.Getenv("WERF_KUBE_AUTH_USERNAME"), "Basic auth username for Kubernetes API (default $WERF_KUBE_AUTH_USERNAME)")
	cmd.Flags().StringVarP(&cmdData.KubeBearerTokenData, "kube-token", "", os.Getenv("WERF_KUBE_TOKEN"), "Kubernetes bearer token used for authentication (default $WERF_KUBE_TOKEN)")
	cmd.Flags().StringVarP(&cmdData.KubeBearerTokenPath, "kube-token-path", "", os.Getenv("WERF_KUBE_TOKEN_PATH"), "Path to file with bearer token for authentication in Kubernetes (default $WERF_KUBE_TOKEN_PATH)")
	if defVal, err := util.GetIntEnvVarDefault("WERF_KUBE_BURST_LIMIT", common.DefaultBurstLimit); err != nil {
		return fmt.Errorf("bad WERF_KUBE_BURST_LIMIT value: %w", err)
	} else {
		cmd.Flags().IntVarP(&cmdData.KubeBurstLimit, "kube-burst-limit", "", defVal, fmt.Sprintf("Kubernetes client burst limit (default $WERF_KUBE_BURST_LIMIT or %d)", common.DefaultBurstLimit))
	}
	SetupKubeConfigBase64(cmdData, cmd)
	SetupLegacyKubeConfigPath(cmdData, cmd)
	cmd.Flags().StringVarP(&cmdData.KubeContextCluster, "kube-context-cluster", "", os.Getenv("WERF_KUBE_CONTEXT_CLUSTER"), "Use cluster from Kubeconfig for current context (default $WERF_KUBE_CONTEXT_CLUSTER)")
	SetupKubeContextCurrent(cmdData, cmd)
	cmd.Flags().StringVarP(&cmdData.KubeContextUser, "kube-context-user", "", os.Getenv("WERF_KUBE_CONTEXT_USER"), "Use user from Kubeconfig for current context (default $WERF_KUBE_CONTEXT_USER)")
	cmd.Flags().StringArrayVarP(&cmdData.KubeImpersonateGroups, "kube-impersonate-group", "", []string{}, "Sets Impersonate-Group headers when authenticating in Kubernetes. Can be also set with $WERF_KUBE_IMPERSONATE_GROUP_* environment variables")
	cmd.Flags().StringVarP(&cmdData.KubeImpersonateUID, "kube-impersonate-uid", "", os.Getenv("WERF_KUBE_IMPERSONATE_UID"), "Sets Impersonate-Uid header when authenticating in Kubernetes (default $WERF_KUBE_IMPERSONATE_UID)")
	cmd.Flags().StringVarP(&cmdData.KubeImpersonateUser, "kube-impersonate-user", "", os.Getenv("WERF_KUBE_IMPERSONATE_USER"), "Sets Impersonate-User header when authenticating in Kubernetes (default $WERF_KUBE_IMPERSONATE_USER)")
	cmd.Flags().StringVarP(&cmdData.KubeProxyURL, "kube-proxy-url", "", os.Getenv("WERF_KUBE_PROXY_URL"), "Proxy URL to use for proxying all requests to Kubernetes API (default $WERF_KUBE_PROXY_URL)")
	if defVal, err := util.GetIntEnvVarDefault("WERF_KUBE_QPS_LIMIT", common.DefaultQPSLimit); err != nil {
		return fmt.Errorf("bad WERF_KUBE_QPS_LIMIT value: %w", err)
	} else {
		cmd.Flags().IntVarP(&cmdData.KubeQPSLimit, "kube-qps-limit", "", defVal, fmt.Sprintf("Kubernetes client QPS limit (default $WERF_KUBE_QPS_LIMIT or %d)", common.DefaultQPSLimit))
	}
	if defVal, err := util.GetDurationEnvVar("WERF_KUBE_REQUEST_TIMEOUT"); err != nil {
		return fmt.Errorf("bad WERF_KUBE_REQUEST_TIMEOUT value: %w", err)
	} else {
		cmd.Flags().DurationVarP(&cmdData.KubeRequestTimeout, "kube-request-timeout", "", defVal, "Timeout for all requests to Kubernetes API (default $WERF_KUBE_REQUEST_TIMEOUT)")
	}
	cmd.Flags().BoolVarP(&cmdData.KubeSkipTLSVerify, "skip-tls-verify-kube", "", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_TLS_VERIFY_KUBE"), "Skip TLS certificate validation when accessing a Kubernetes cluster (default $WERF_SKIP_TLS_VERIFY_KUBE)")
	cmd.Flags().StringVarP(&cmdData.KubeTLSCAData, "kube-ca-data", "", os.Getenv("WERF_KUBE_CA_DATA"), "Pass Kubernetes API server TLS CA data (default $WERF_KUBE_CA_DATA)")
	cmd.Flags().StringVarP(&cmdData.KubeTLSCAPath, "kube-ca-path", "", os.Getenv("WERF_KUBE_CA_PATH"), "Kubernetes API server CA path (default $WERF_KUBE_CA_PATH)")
	cmd.Flags().StringVarP(&cmdData.KubeTLSClientCertData, "kube-cert-data", "", os.Getenv("WERF_KUBE_CERT_DATA"), "Pass PEM-encoded TLS client cert for connecting to Kubernetes API (default $WERF_KUBE_CERT_DATA)")
	cmd.Flags().StringVarP(&cmdData.KubeTLSClientCertPath, "kube-cert", "", os.Getenv("WERF_KUBE_CERT"), "Path to PEM-encoded TLS client cert for connecting to Kubernetes API (default $WERF_KUBE_CERT")
	cmd.Flags().StringVarP(&cmdData.KubeTLSClientKeyData, "kube-key-data", "", os.Getenv("WERF_KUBE_KEY_DATA"), "Pass PEM-encoded TLS client key for connecting to Kubernetes API (default $WERF_KUBE_KEY_DATA)")
	cmd.Flags().StringVarP(&cmdData.KubeTLSClientKeyPath, "kube-key", "", os.Getenv("WERF_KUBE_KEY"), "Path to PEM-encoded TLS client key for connecting to Kubernetes API (default $WERF_KUBE_KEY)")
	cmd.Flags().StringVarP(&cmdData.KubeTLSServerName, "kube-tls-server", "", os.Getenv("WERF_KUBE_TLS_SERVER"), "Server name to use for Kubernetes API server certificate validation. If it is not provided, the hostname used to contact the server is used (default $WERF_KUBE_TLS_SERVER)")

	return nil
}

func SetupChartRepoConnectionFlags(cmdData *CmdData, cmd *cobra.Command) error {
	SetupChartRepoInsecure(cmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.ChartRepoSkipTLSVerify, "skip-tls-verify-helm-dependencies", "", util.GetBoolEnvironmentDefaultFalse("WERF_SKIP_TLS_VERIFY_HELM_DEPENDENCIES"), "Skip TLS certificate validation when accessing a Helm charts repository (default $WERF_SKIP_TLS_VERIFY_HELM_DEPENDENCIES)")

	return nil
}

func SetupValuesFlags(cmdData *CmdData, cmd *cobra.Command) error {
	cmd.Flags().BoolVarP(&cmdData.DefaultValuesDisable, "disable-default-values", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_DEFAULT_VALUES"), `Do not use values from the default .helm/values.yaml file (default $WERF_DISABLE_DEFAULT_VALUES or false)`)
	cmd.Flags().StringArrayVarP(&cmdData.RootSetJSON, "set-root-json", "", []string{}, `Set new keys in arbitrary things in the global root context ("$"), where the key is the value path and the value is JSON. This is meant to be generated inside the program, so use --set-json instead, unless you REALLY know what you are doing. Can specify multiple or separate values with commas: key1=val1,key2=val2.
Also, can be defined with $WERF_SET_ROOT_JSON_* (e.g. $WERF_SET_ROOT_JSON_1=key1=val1, $WERF_SET_ROOT_JSON_2=key2=val2)`)
	cmd.Flags().StringArrayVarP(&cmdData.RuntimeSetJSON, "set-runtime-json", "", []string{}, `Set new keys in $.Runtime, where the key is the value path and the value is JSON. This is meant to be generated inside the program, so use --set-json instead, unless you know what you are doing. Can specify multiple or separate values with commas: key1=val1,key2=val2.
Also, can be defined with $WERF_SET_RUNTIME_JSON_* (e.g. $WERF_SET_RUNTIME_JSON_1=key1=val1, $WERF_SET_RUNTIME_JSON_2=key2=val2)`)
	cmd.Flags().StringArrayVarP(&cmdData.ValuesFiles, "values", "", []string{}, `Specify helm values in a YAML file or a URL (can specify multiple). Also, can be defined with $WERF_VALUES_* (e.g. $WERF_VALUES_1=.helm/values_1.yaml, $WERF_VALUES_2=.helm/values_2.yaml)`)
	cmd.Flags().StringArrayVarP(&cmdData.ValuesSet, "set", "", []string{}, `Set helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET_* (e.g. $WERF_SET_1=key1=val1, $WERF_SET_2=key2=val2)`)
	cmd.Flags().StringArrayVarP(&cmdData.ValuesSetFile, "set-file", "", []string{}, `Set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2).
Also, can be defined with $WERF_SET_FILE_* (e.g. $WERF_SET_FILE_1=key1=path1, $WERF_SET_FILE_2=key2=val2)`)
	cmd.Flags().StringArrayVarP(&cmdData.ValuesSetJSON, "set-json", "", []string{}, `Set new values, where the key is the value path and the value is JSON (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET_JSON_* (e.g. $WERF_SET_JSON_1=key1=val1, $WERF_SET_JSON_2=key2=val2)`)
	cmd.Flags().StringArrayVarP(&cmdData.ValuesSetLiteral, "set-literal", "", []string{}, `Set new values, where the key is the value path and the value is the value. The value will always become a literal string (can specify multiple or separate values with commas: key1=val1,key2=val2).)
Also, can be defined with $WERF_SET_LITERAL_* (e.g. $WERF_SET_LITERAL_1=key1=val1, $WERF_SET_LITERAL_2=key2=val2)`)
	cmd.Flags().StringArrayVarP(&cmdData.ValuesSetString, "set-string", "", []string{}, `Set STRING helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET_STRING_* (e.g. $WERF_SET_STRING_1=key1=val1, $WERF_SET_STRING_2=key2=val2)`)

	return nil
}

func SetupSecretValuesFlags(cmdData *CmdData, cmd *cobra.Command) error {
	cmd.Flags().BoolVarP(&cmdData.DefaultSecretValuesDisable, "disable-default-secret-values", "", util.GetBoolEnvironmentDefaultFalse("WERF_DISABLE_DEFAULT_SECRET_VALUES"), `Do not use secret values from the default .helm/secret-values.yaml file (default $WERF_DISABLE_DEFAULT_SECRET_VALUES or false)`)
	cmd.Flags().StringVarP(&cmdData.SecretKey, "secret-key", "", os.Getenv("WERF_SECRET_KEY"), "Secret key (default $WERF_SECRET_KEY)")
	cmd.Flags().BoolVarP(&cmdData.SecretKeyIgnore, "ignore-secret-key", "", util.GetBoolEnvironmentDefaultFalse("WERF_IGNORE_SECRET_KEY"), "Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)")
	cmd.Flags().StringArrayVarP(&cmdData.SecretValuesFiles, "secret-values", "", []string{}, `Specify helm secret values in a YAML file (can specify multiple). Also, can be defined with $WERF_SECRET_VALUES_* (e.g. $WERF_SECRET_VALUES_ENV=.helm/secret_values_test.yaml, $WERF_SECRET_VALUES_DB=.helm/secret_values_db.yaml)`)

	return nil
}

func SetupTrackingFlags(cmdData *CmdData, cmd *cobra.Command) error {
	SetupNoFinalTrackingFlag(cmdData, cmd)
	cmd.PersistentFlags().BoolVarP(&cmdData.NoPodLogs, "no-pod-logs", "", false, "Disable Pod logs collection and printing (default $WERF_NO_POD_LOGS or false)")
	if err := SetupLegacyProgressTablePrintInterval(cmdData, cmd); err != nil {
		return err
	}
	if defVal, err := util.GetIntEnvVar("WERF_TIMEOUT"); err != nil {
		return fmt.Errorf("bad WERF_TIMEOUT value: %w", err)
	} else {
		var def int
		if defVal != nil {
			def = int(*defVal)
		}

		cmd.Flags().IntVarP(&cmdData.LegacyTrackTimeout, "timeout", "t", def, "Resources tracking timeout in seconds ($WERF_TIMEOUT by default)")
	}

	StubSetupHooksStatusProgressPeriod(cmdData, cmd)

	return nil
}

func SetupResourceValidationFlags(cmdData *CmdData, cmd *cobra.Command) error {
	if !featgate.FeatGateResourceValidation.Enabled() {
		return nil
	}

	kubeVersion := os.Getenv("WERF_RESOURCE_VALIDATION_KUBE_VERSION")
	if kubeVersion == "" {
		kubeVersion = common.DefaultResourceValidationKubeVersion
	}

	defaultValidationCacheLifetime := common.DefaultResourceValidationCacheLifetime

	if os.Getenv("WERF_RESOURCE_VALIDATION_CACHE_LIFETIME") != "" {
		var err error

		defaultValidationCacheLifetime, err = util.GetDurationEnvVar("WERF_RESOURCE_VALIDATION_CACHE_LIFETIME")
		if err != nil {
			return fmt.Errorf("bad WERF_RESOURCE_VALIDATION_CACHE_LIFETIME value: %w", err)
		}
	}

	validationSchemas := util.PredefinedValuesByEnvNamePrefix("WERF_RESOURCE_VALIDATION_SCHEMA_")
	if len(validationSchemas) == 0 {
		validationSchemas = common.DefaultResourceValidationSchema
	}

	cmd.Flags().BoolVarP(&cmdData.NoResourceValidation, "no-resource-validation", "", util.GetBoolEnvironmentDefaultFalse("WERF_NO_RESOURCE_VALIDATION"), "Disable resource validation (default $WERF_NO_RESOURCE_VALIDATION)")
	cmd.Flags().BoolVarP(&cmdData.LocalResourceValidation, "local-resource-validation", "", util.GetBoolEnvironmentDefaultFalse("WERF_LOCAL_RESOURCE_VALIDATION"), "Do not use external json schema sources (default $WERF_LOCAL_RESOURCE_VALIDATION)")
	cmd.Flags().StringVarP(&cmdData.ValidationKubeVersion, "resource-validation-kube-version", "", kubeVersion, "Kubernetes schemas version to use during resource validation. Also can be defined by $WERF_RESOURCE_VALIDATION_KUBE_VERSION")
	cmd.Flags().StringArrayVarP(&cmdData.ValidationSkip, "resource-validation-skip", "", []string{}, "Skip resource validation for resources with specified attributes (can specify multiple). Format: key1=value1,key2=value2. Supported keys: group, version, kind, name, namespace. Example: kind=Deployment,name=my-app. Also, can be defined with $WERF_RESOURCE_VALIDATION_SKIP_* (e.g. $WERF_RESOURCE_VALIDATION_SKIP_1=kind=Deployment,name=my-app)")
	cmd.Flags().StringArrayVarP(&cmdData.ValidationSchemas, "resource-validation-schema", "", validationSchemas, "Default json schema sources to validate resources. Must be a valid go template defining a http(s) URL, or an absolute path on local file system. Also, can be defined with $WERF_RESOURCE_VALIDATION_SCHEMA_* (eg. $WERF_RESOURCE_VALIDATION_SCHEMA_1='https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json')")
	cmd.Flags().StringArrayVarP(&cmdData.ValidationExtraSchemas, "resource-validation-extra-schema", "", []string{}, "Extra json schema sources to validate resources (preferred over default sources). Must be a valid go template defining a http(s) URL, or an absolute path on local file system. Also, can be defined with $WERF_RESOURCE_VALIDATION_EXTRA_SCHEMA_* (eg. $WERF_RESOURCE_VALIDATION_EXTRA_SCHEMA_1='https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json')")
	cmd.Flags().DurationVarP(&cmdData.ValidationSchemaCacheLifetime, "resource-validation-cache-lifetime", "", defaultValidationCacheLifetime, "How long local schema cache will be valid. Also can be defined by $WERF_RESOURCE_VALIDATION_CACHE_LIFETIME")

	return nil
}

func SetupKubeConfigBase64(cmdData *CmdData, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cmdData.KubeConfigBase64, "kube-config-base64", "", GetFirstExistingKubeConfigBase64EnvVar(), "Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)")
}

func SetupNoFinalTrackingFlag(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.NoFinalTracking, "no-final-tracking", "", util.GetBoolEnvironmentDefaultFalse("WERF_NO_FINAL_TRACKING"), `By default disable tracking operations that have no create/update/delete resource operations after them, which are most tracking operations, to speed up the release (default $WERF_NO_FINAL_TRACKING)`)
}

func SetupLegacyKubeConfigPath(cmdData *CmdData, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cmdData.LegacyKubeConfigPath, "kube-config", "", "", "Kubernetes config file path (default $WERF_KUBE_CONFIG, or $WERF_KUBECONFIG, or $KUBECONFIG)")
	cmdData.LegacyKubeConfigPathsMergeList = filepath.SplitList(GetFirstExistingKubeConfigEnvVar())
}

func SetupKubeContextCurrent(cmdData *CmdData, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cmdData.KubeContextCurrent, "kube-context", "", os.Getenv("WERF_KUBE_CONTEXT"), "Kubernetes config context (default $WERF_KUBE_CONTEXT)")
}

func SetupLegacyProgressTablePrintInterval(cmdData *CmdData, cmd *cobra.Command) error {
	if defVal, err := util.GetFirstExistingEnvVarAsInt("WERF_STATUS_PROGRESS_PERIOD_SECONDS", "WERF_STATUS_PROGRESS_PERIOD"); err != nil {
		return fmt.Errorf("bad WERF_STATUS_PROGRESS_PERIOD_SECONDS or WERF_STATUS_PROGRESS_PERIOD value: %w", err)
	} else {
		if defVal == nil {
			defVal = lo.ToPtr(int(common.DefaultProgressPrintInterval.Seconds()))
		}

		cmd.PersistentFlags().IntVarP(&cmdData.LegacyProgressTablePrintInterval, "status-progress-period", "", *defVal, "Status progress period in seconds. Set -1 to stop showing status progress. Defaults to $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds")
	}

	return nil
}

func SetupChartRepoInsecure(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&cmdData.ChartRepoInsecure, "insecure-helm-dependencies", "", util.GetBoolEnvironmentDefaultFalse("WERF_INSECURE_HELM_DEPENDENCIES"), "Allow insecure oci registries to be used in the Chart.yaml dependencies configuration (default $WERF_INSECURE_HELM_DEPENDENCIES)")
}

// TODO(major): remove
func StubSetupInsecureHelmDependencies(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().BoolVar(lo.ToPtr(false), "insecure-helm-dependencies", false, "No-op")
}

// TODO(major): remove
func StubSetupStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmd.PersistentFlags().IntVarP(lo.ToPtr(0), "status-progress-period", "", 0, "No-op")
}

// TODO(major): remove
func StubSetupHooksStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmd.PersistentFlags().Int64VarP(lo.ToPtr(int64(0)), "hooks-status-progress-period", "", 0, "No-op")
}

// TODO(major): remove
func StubSetupTrackTimeout(cmdData *CmdData, cmd *cobra.Command) {
	cmd.Flags().IntVarP(lo.ToPtr(0), "timeout", "t", 0, "No-op")
}

func HasKubeConfig(cmdData *CmdData) bool {
	return cmdData.LegacyKubeConfigPath != "" ||
		cmdData.KubeConfigBase64 != "" ||
		len(cmdData.LegacyKubeConfigPathsMergeList) > 0 ||
		cmdData.KubeBearerTokenData != "" ||
		cmdData.KubeBearerTokenPath != "" ||
		cmdData.KubeAPIServerAddress != ""
}
