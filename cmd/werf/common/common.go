package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

type CmdData struct {
	ProjectName        *string
	Dir                *string
	ConfigPath         *string
	ConfigTemplatesDir *string
	TmpDir             *string
	HomeDir            *string
	SSHKeys            *[]string

	HelmChartDir                     *string
	Environment                      *string
	Release                          *string
	Namespace                        *string
	AddAnnotations                   *[]string
	AddLabels                        *[]string
	KubeContext                      *string
	KubeConfig                       *string
	KubeConfigBase64                 *string
	StatusProgressPeriodSeconds      *int64
	HooksStatusProgressPeriodSeconds *int64
	ReleasesHistoryMax               *int

	Set             *[]string
	SetString       *[]string
	Values          *[]string
	SetFile         *[]string
	SecretValues    *[]string
	IgnoreSecretKey *bool

	CommonRepoData         *RepoData
	StagesStorage          *string
	SecondaryStagesStorage *[]string

	SkipBuild *bool
	StubTags  *bool

	Synchronization    *string
	Parallel           *bool
	ParallelTasksLimit *int64

	DockerConfig          *string
	InsecureRegistry      *bool
	SkipTlsVerifyRegistry *bool
	DryRun                *bool
	DisableDeterminism    *bool

	KeepStagesBuiltWithinLastNHours *uint64
	WithoutKube                     *bool

	IntrospectBeforeError *bool
	IntrospectAfterError  *bool
	StagesToIntrospect    *[]string

	Follow *bool

	LogDebug         *bool
	LogPretty        *bool
	LogVerbose       *bool
	LogQuiet         *bool
	LogColorMode     *string
	LogProjectDir    *bool
	LogTerminalWidth *int64

	ReportPath   *string
	ReportFormat *string

	VirtualMerge           *bool
	VirtualMergeFromCommit *string
	VirtualMergeIntoCommit *string

	ScanContextNamespaceOnly *bool
}

const (
	CleaningCommandsForceOptionDescription = "First remove containers that use werf docker images which are going to be deleted"
	StubRepoAddress                        = "stub/repository"
	StubTag                                = "TAG"
	DefaultBuildParallelTasksLimit         = 5
	DefaultCleanupParallelTasksLimit       = 10
)

func GetLongCommandDescription(text string) string {
	return logboek.FitText(text, types.FitTextOptions{MaxWidth: 100})
}

func SetupProjectName(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ProjectName = new(string)
	cmd.Flags().StringVarP(cmdData.ProjectName, "project-name", "N", os.Getenv("WERF_PROJECT_NAME"), "Use custom project name (default $WERF_PROJECT_NAME)")
}

func SetupDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Dir = new(string)
	cmd.Flags().StringVarP(cmdData.Dir, "dir", "", os.Getenv("WERF_DIR"), "Use custom working directory (default $WERF_DIR or current directory)")
}

func SetupConfigPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigPath = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigPath, "config", "", os.Getenv("WERF_CONFIG"), `Use custom configuration file (default $WERF_CONFIG or werf.yaml in working directory)`)
}

func SetupConfigTemplatesDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ConfigTemplatesDir = new(string)
	cmd.Flags().StringVarP(cmdData.ConfigTemplatesDir, "config-templates-dir", "", os.Getenv("WERF_CONFIG_TEMPLATES_DIR"), `Custom configuration templates directory (default $WERF_CONFIG_TEMPLATES_DIR or .werf in working directory)`)
}

func SetupTmpDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.TmpDir = new(string)
	cmd.Flags().StringVarP(cmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (default $WERF_TMP_DIR or system tmp dir)")
}

func SetupDisableDeterminism(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DisableDeterminism = new(bool)
	cmd.Flags().BoolVarP(cmdData.DisableDeterminism, "disable-determinism", "", GetBoolEnvironmentDefaultFalse("WERF_DISABLE_DETERMINISM"), "Disable werf deterministic mode (more info https://werf.io/documentation/advanced/configuration/determinism.html, default $WERF_DISABLE_DETERMINISM)")
}

func SetupHomeDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HomeDir = new(string)
	cmd.Flags().StringVarP(cmdData.HomeDir, "home-dir", "", "", "Use specified dir to store werf cache files and dirs (default $WERF_HOME or ~/.werf)")
}

func SetupSSHKey(cmdData *CmdData, cmd *cobra.Command) {
	sshKeys := predefinedValuesByEnvNamePrefix("WERF_SSH_KEY")

	cmdData.SSHKeys = &sshKeys
	cmd.Flags().StringArrayVarP(cmdData.SSHKeys, "ssh-key", "", sshKeys, `Use only specific ssh key(s).
Can be specified with $WERF_SSH_KEY* (e.g. $WERF_SSH_KEY_REPO=~/.ssh/repo_rsa", $WERF_SSH_KEY_NODEJS=~/.ssh/nodejs_rsa").
Defaults to $WERF_SSH_KEY*, system ssh-agent or ~/.ssh/{id_rsa|id_dsa}, see https://werf.io/documentation/reference/toolbox/ssh.html`)
}

func SetupReportPath(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ReportPath = new(string)
	cmd.Flags().StringVarP(cmdData.ReportPath, "report-path", "", os.Getenv("WERF_REPORT_PATH"), "Report save path ($WERF_REPORT_PATH by default)")
}

func SetupReportFormat(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ReportFormat = new(string)
	cmd.Flags().StringVarP(cmdData.ReportFormat, "report-format", "", string(build.ReportJSON), fmt.Sprintf(`Report format: %[1]s or %[2]s (%[1]s or $WERF_REPORT_FORMAT by default)
%[1]s:
	{
	  "Images": {
		"<WERF_IMAGE_NAME>": {
			"WerfImageName": "<WERF_IMAGE_NAME>",
			"DockerRepo": "<REPO>",
			"DockerTag": "<TAG>"
			"DockerImageName": "<REPO>:<TAG>",
			"DockerImageID": "<SHA256>",
		},
		...
	  }
	}
%[2]s:
	WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME=<REPO>:<TAG>
	...
<FORMATTED_WERF_IMAGE_NAME> is werf image name from werf.yaml modified according to the following rules:
- all characters are uppercase (app -> APP);
- charset /- is replaced with _ (dev/app-frontend -> DEV_APP_FRONTEND)`, string(build.ReportJSON), string(build.ReportEnvFile)))
}

func GetReportFormat(cmdData *CmdData) (build.ReportFormat, error) {
	switch format := build.ReportFormat(*cmdData.ReportFormat); format {
	case build.ReportJSON, build.ReportEnvFile:
		return format, nil
	default:
		return "", fmt.Errorf("bad --report-format given %q, expected: \"%s\"", format, strings.Join([]string{string(build.ReportJSON), string(build.ReportEnvFile)}, "\", \""))
	}
}

func SetupWithoutKube(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.WithoutKube = new(bool)
	cmd.Flags().BoolVarP(cmdData.WithoutKube, "without-kube", "", GetBoolEnvironmentDefaultFalse("WERF_WITHOUT_KUBE"), "Do not skip deployed Kubernetes images (default $WERF_WITHOUT_KUBE)")
}

func SetupKeepStagesBuiltWithinLastNHours(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KeepStagesBuiltWithinLastNHours = new(uint64)

	envValue, err := getUint64EnvVar("WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	var defaultValue uint64
	if envValue != nil {
		defaultValue = *envValue
	} else {
		defaultValue = 2
	}

	cmd.Flags().Uint64VarP(cmdData.KeepStagesBuiltWithinLastNHours, "keep-stages-built-within-last-n-hours", "", defaultValue, "Keep stages that were built within last hours (default $WERF_KEEP_STAGES_BUILT_WITHIN_LAST_N_HOURS or 2)")
}

func predefinedValuesByEnvNamePrefix(envNamePrefix string, envNamePrefixesToExcept ...string) []string {
	var result []string

environLoop:
	for _, keyValue := range os.Environ() {
		parts := strings.SplitN(keyValue, "=", 2)
		if strings.HasPrefix(parts[0], envNamePrefix) {
			for _, exceptEnvNamePrefix := range envNamePrefixesToExcept {
				if strings.HasPrefix(parts[0], exceptEnvNamePrefix) {
					continue environLoop
				}
			}

			result = append(result, parts[1])
		}
	}

	return result
}

func SetupEnvironment(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Environment = new(string)
	cmd.Flags().StringVarP(cmdData.Environment, "env", "", os.Getenv("WERF_ENV"), "Use specified environment (default $WERF_ENV)")
}

func SetupRelease(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Release = new(string)
	cmd.Flags().StringVarP(cmdData.Release, "release", "", os.Getenv("WERF_RELEASE"), "Use specified Helm release name (default [[ project ]]-[[ env ]] template or deploy.helmRelease custom template from werf.yaml or $WERF_RELEASE)")
}

func SetupNamespace(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Namespace = new(string)
	cmd.Flags().StringVarP(cmdData.Namespace, "namespace", "", os.Getenv("WERF_NAMESPACE"), "Use specified Kubernetes namespace (default [[ project ]]-[[ env ]] template or deploy.namespace custom template from werf.yaml or $WERF_NAMESPACE)")
}

func SetupAddAnnotations(cmdData *CmdData, cmd *cobra.Command) {
	addAnnotations := predefinedValuesByEnvNamePrefix("WERF_ADD_ANNOTATION")

	cmdData.AddAnnotations = &addAnnotations
	cmd.Flags().StringArrayVarP(cmdData.AddAnnotations, "add-annotation", "", addAnnotations, `Add annotation to deploying resources (can specify multiple).
Format: annoName=annoValue.
Also, can be specified with $WERF_ADD_ANNOTATION* (e.g. $WERF_ADD_ANNOTATION_1=annoName1=annoValue1", $WERF_ADD_ANNOTATION_2=annoName2=annoValue2")`)
}

func SetupAddLabels(cmdData *CmdData, cmd *cobra.Command) {
	addLabels := predefinedValuesByEnvNamePrefix("WERF_ADD_LABEL")

	cmdData.AddLabels = &addLabels
	cmd.Flags().StringArrayVarP(cmdData.AddLabels, "add-label", "", addLabels, `Add label to deploying resources (can specify multiple).
Format: labelName=labelValue.
Also, can be specified with $WERF_ADD_LABEL* (e.g. $WERF_ADD_LABEL_1=labelName1=labelValue1", $WERF_ADD_LABEL_2=labelName2=labelValue2")`)
}

func SetupKubeContext(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeContext = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.KubeContext, "kube-context", "", os.Getenv("WERF_KUBE_CONTEXT"), "Kubernetes config context (default $WERF_KUBE_CONTEXT)")
}

func SetupKubeConfig(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeConfig = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.KubeConfig, "kube-config", "", getFirstExistingEnvVarAsString("WERF_KUBE_CONFIG", "WERF_KUBECONFIG", "KUBECONFIG"), "Kubernetes config file path (default $WERF_KUBE_CONFIG or $WERF_KUBECONFIG or $KUBECONFIG)")
}

func SetupKubeConfigBase64(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.KubeConfigBase64 = new(string)
	cmd.PersistentFlags().StringVarP(cmdData.KubeConfigBase64, "kube-config-base64", "", getFirstExistingEnvVarAsString("WERF_KUBE_CONFIG_BASE64", "WERF_KUBECONFIG_BASE64", "KUBECONFIG_BASE64"), "Kubernetes config data as base64 string (default $WERF_KUBE_CONFIG_BASE64 or $WERF_KUBECONFIG_BASE64 or $KUBECONFIG_BASE64)")
}

func getFirstExistingEnvVarAsString(envNames ...string) string {
	for _, envName := range envNames {
		if v := os.Getenv(envName); v != "" {
			return v
		}
	}

	return ""
}

func SetupCommonRepoData(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.CommonRepoData = &RepoData{IsCommon: true}

	SetupImplementationForRepoData(cmdData.CommonRepoData, cmd, "repo-implementation", []string{"WERF_REPO_IMPLEMENTATION"})
	SetupDockerHubUsernameForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-username", []string{"WERF_REPO_DOCKER_HUB_USERNAME"})
	SetupDockerHubPasswordForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-password", []string{"WERF_REPO_DOCKER_HUB_PASSWORD"})
	SetupDockerHubTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-docker-hub-token", []string{"WERF_REPO_DOCKER_HUB_TOKEN"})
	SetupGithubTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-github-token", []string{"WERF_REPO_GITHUB_TOKEN"})
	SetupHarborUsernameForRepoData(cmdData.CommonRepoData, cmd, "repo-harbor-username", []string{"WERF_REPO_HARBOR_USERNAME"})
	SetupHarborPasswordForRepoData(cmdData.CommonRepoData, cmd, "repo-harbor-password", []string{"WERF_REPO_HARBOR_PASSWORD"})
	SetupQuayTokenForRepoData(cmdData.CommonRepoData, cmd, "repo-quay-token", []string{"WERF_REPO_QUAY_TOKEN"})
}

func SetupSecondaryStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	secondaryStagesStorage := predefinedValuesByEnvNamePrefix("WERF_SECONDARY_REPO")
	cmdData.SecondaryStagesStorage = &secondaryStagesStorage
	cmd.Flags().StringArrayVarP(cmdData.SecondaryStagesStorage, "secondary-repo", "", secondaryStagesStorage, "Specify one or multiple secondary read-only repo with images that will be used as a cache")
}

func SetupStagesStorageOptions(cmdData *CmdData, cmd *cobra.Command) {
	SetupInsecureRegistry(cmdData, cmd)
	SetupSkipTlsVerifyRegistry(cmdData, cmd)
	SetupCommonRepoData(cmdData, cmd)
	setupStagesStorage(cmdData, cmd)
}

func setupStagesStorage(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.StagesStorage, "repo", "", os.Getenv("WERF_REPO"), fmt.Sprintf("Docker Repo to store stages (default $WERF_REPO)"))
}

func SetupStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StatusProgressPeriodSeconds = new(int64)
	SetupStatusProgressPeriodP(cmdData.StatusProgressPeriodSeconds, cmd)
}

func SetupStatusProgressPeriodP(destination *int64, cmd *cobra.Command) {
	cmd.PersistentFlags().Int64VarP(
		destination,
		"status-progress-period",
		"",
		*statusProgressPeriodDefaultValue(),
		"Status progress period in seconds. Set -1 to stop showing status progress. Defaults to $WERF_STATUS_PROGRESS_PERIOD_SECONDS or 5 seconds",
	)
}

func SetupReleasesHistoryMax(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.ReleasesHistoryMax = new(int)

	defaultValueP, err := getIntEnvVar("WERF_RELEASES_HISTORY_MAX")
	if err != nil {
		TerminateWithError(fmt.Sprintf("bad WERF_RELEASES_HISTORY_MAX value: %s", err), 1)
	}

	var defaultValue int
	if defaultValueP != nil {
		defaultValue = int(*defaultValueP)
	}

	cmd.Flags().IntVarP(
		cmdData.ReleasesHistoryMax,
		"releases-history-max",
		"",
		defaultValue,
		"Max releases to keep in release storage. Can be set by environment variable $WERF_RELEASES_HISTORY_MAX. By default werf keeps all releases.",
	)
}

func statusProgressPeriodDefaultValue() *int64 {
	defaultValue := int64(5)

	v, err := getIntEnvVar("WERF_STATUS_PROGRESS_PERIOD_SECONDS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	if v == nil {
		return &defaultValue
	} else {
		return v
	}
}

func SetupHooksStatusProgressPeriod(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.HooksStatusProgressPeriodSeconds = new(int64)
	SetupHooksStatusProgressPeriodP(cmdData.HooksStatusProgressPeriodSeconds, cmd)
}

func SetupHooksStatusProgressPeriodP(destination *int64, cmd *cobra.Command) {
	cmd.PersistentFlags().Int64VarP(
		destination,
		"hooks-status-progress-period",
		"",
		*hooksStatusProgressPeriodDefaultValue(),
		"Hooks status progress period in seconds. Set 0 to stop showing hooks status progress. Defaults to $WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS or status progress period value",
	)
}

func hooksStatusProgressPeriodDefaultValue() *int64 {
	defaultValue := statusProgressPeriodDefaultValue()

	v, err := getIntEnvVar("WERF_HOOKS_STATUS_PROGRESS_PERIOD_SECONDS")
	if err != nil {
		TerminateWithError(err.Error(), 1)
	}

	if v == nil {
		return defaultValue
	} else {
		return v
	}
}

func SetupInsecureRegistry(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.InsecureRegistry != nil {
		return
	}

	cmdData.InsecureRegistry = new(bool)
	cmd.Flags().BoolVarP(cmdData.InsecureRegistry, "insecure-registry", "", GetBoolEnvironmentDefaultFalse("WERF_INSECURE_REGISTRY"), "Use plain HTTP requests when accessing a registry (default $WERF_INSECURE_REGISTRY)")
}

func SetupSkipTlsVerifyRegistry(cmdData *CmdData, cmd *cobra.Command) {
	if cmdData.SkipTlsVerifyRegistry != nil {
		return
	}

	cmdData.SkipTlsVerifyRegistry = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipTlsVerifyRegistry, "skip-tls-verify-registry", "", GetBoolEnvironmentDefaultFalse("WERF_SKIP_TLS_VERIFY_REGISTRY"), "Skip TLS certificate validation when accessing a registry (default $WERF_SKIP_TLS_VERIFY_REGISTRY)")
}

func SetupDryRun(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.DryRun = new(bool)
	cmd.Flags().BoolVarP(cmdData.DryRun, "dry-run", "", GetBoolEnvironmentDefaultFalse("WERF_DRY_RUN"), "Indicate what the command would do without actually doing that (default $WERF_DRY_RUN)")
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
	setupLogDebug(cmdData, cmd)
	setupLogVerbose(cmdData, cmd)
	setupLogQuiet(cmdData, cmd, false)
	setupLogColor(cmdData, cmd)
	setupLogPretty(cmdData, cmd)
	setupTerminalWidth(cmdData, cmd)
}

func SetupLogOptionsDefaultQuiet(cmdData *CmdData, cmd *cobra.Command) {
	setupLogDebug(cmdData, cmd)
	setupLogVerbose(cmdData, cmd)
	setupLogQuiet(cmdData, cmd, true)
	setupLogColor(cmdData, cmd)
	setupLogPretty(cmdData, cmd)
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
			defaultValue = GetBoolEnvironmentDefaultFalse(envName)
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

	var defaultValue = isDefaultQuiet

	for _, envName := range []string{
		"WERF_LOG_QUIET",
		"WERF_QUIET",
	} {
		if os.Getenv(envName) != "" {
			if defaultValue {
				defaultValue = GetBoolEnvironmentDefaultTrue(envName)
			} else {
				defaultValue = GetBoolEnvironmentDefaultFalse(envName)
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
			defaultValue = GetBoolEnvironmentDefaultFalse(envName)
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
		defaultValue = GetBoolEnvironmentDefaultFalse("WERF_LOG_PRETTY")
	} else {
		defaultValue = true
	}

	cmd.PersistentFlags().BoolVarP(cmdData.LogPretty, "log-pretty", "", defaultValue, `Enable emojis, auto line wrapping and log process border (default $WERF_LOG_PRETTY or true).`)
}

func setupTerminalWidth(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogTerminalWidth = new(int64)
	cmd.PersistentFlags().Int64VarP(cmdData.LogTerminalWidth, "log-terminal-width", "", -1, fmt.Sprintf(`Set log terminal width.
Defaults to:
* $WERF_LOG_TERMINAL_WIDTH
* interactive terminal width or %d`, 140))
}

func SetupSet(cmdData *CmdData, cmd *cobra.Command) {
	set := predefinedValuesByEnvNamePrefix("WERF_SET", "WERF_SET_STRING")

	cmdData.Set = &set
	cmd.Flags().StringArrayVarP(cmdData.Set, "set", "", set, `Set helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET* (e.g. $WERF_SET_1=key1=val1, $WERF_SET_2=key2=val2)`)
}

func SetupSetString(cmdData *CmdData, cmd *cobra.Command) {
	setString := predefinedValuesByEnvNamePrefix("WERF_SET_STRING")

	cmdData.SetString = &setString
	cmd.Flags().StringArrayVarP(cmdData.SetString, "set-string", "", setString, `Set STRING helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2).
Also, can be defined with $WERF_SET_STRING* (e.g. $WERF_SET_STRING_1=key1=val1, $WERF_SET_STRING_2=key2=val2)`)
}

func SetupValues(cmdData *CmdData, cmd *cobra.Command) {
	values := predefinedValuesByEnvNamePrefix("WERF_VALUES")

	cmdData.Values = &values
	cmd.Flags().StringArrayVarP(cmdData.Values, "values", "", values, `Specify helm values in a YAML file or a URL (can specify multiple).
Also, can be defined with $WERF_VALUES* (e.g. $WERF_VALUES_ENV=.helm/values_test.yaml, $WERF_VALUES_DB=.helm/values_db.yaml)`)
}

func SetupSetFile(cmdData *CmdData, cmd *cobra.Command) {
	setFile := predefinedValuesByEnvNamePrefix("WERF_SET_FILE")

	cmdData.SetFile = &setFile
	cmd.Flags().StringArrayVarP(cmdData.SetFile, "set-file", "", setFile, `Set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2).
Also, can be defined with $WERF_SET_FILE* (e.g. $WERF_SET_FILE_1=key1=path1, $WERF_SET_FILE_2=key2=val2)`)
}

func SetupSecretValues(cmdData *CmdData, cmd *cobra.Command) {
	secretValues := predefinedValuesByEnvNamePrefix("WERF_SECRET_VALUES")

	cmdData.SecretValues = &secretValues
	cmd.Flags().StringArrayVarP(cmdData.SecretValues, "secret-values", "", secretValues, `Specify helm secret values in a YAML file (can specify multiple).
Also, can be defined with $WERF_SECRET_VALUES* (e.g. $WERF_SECRET_VALUES_ENV=.helm/secret_values_test.yaml, $WERF_SECRET_VALUES=.helm/secret_values_db.yaml)`)
}

func SetupIgnoreSecretKey(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IgnoreSecretKey = new(bool)
	cmd.Flags().BoolVarP(cmdData.IgnoreSecretKey, "ignore-secret-key", "", GetBoolEnvironmentDefaultFalse("WERF_IGNORE_SECRET_KEY"), "Disable secrets decryption (default $WERF_IGNORE_SECRET_KEY)")
}

func SetupParallelOptions(cmdData *CmdData, cmd *cobra.Command, defaultValue int64) {
	SetupParallel(cmdData, cmd)
	SetupParallelTasksLimit(cmdData, cmd, defaultValue)
}

func SetupParallel(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Parallel = new(bool)
	cmd.Flags().BoolVarP(cmdData.Parallel, "parallel", "p", GetBoolEnvironmentDefaultTrue("WERF_PARALLEL"), "Run in parallel (default $WERF_PARALLEL)")
}

func SetupParallelTasksLimit(cmdData *CmdData, cmd *cobra.Command, defaultValue int64) {
	cmdData.ParallelTasksLimit = new(int64)
	cmd.Flags().Int64VarP(cmdData.ParallelTasksLimit, "parallel-tasks-limit", "", defaultValue, "Parallel tasks limit, set -1 to remove the limitation (default $WERF_PARALLEL_TASKS_LIMIT or 5)")
}

func SetupLogProjectDir(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.LogProjectDir = new(bool)
	cmd.Flags().BoolVarP(cmdData.LogProjectDir, "log-project-dir", "", GetBoolEnvironmentDefaultFalse("WERF_LOG_PROJECT_DIR"), `Print current project directory path (default $WERF_LOG_PROJECT_DIR)`)
}

func SetupIntrospectAfterError(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IntrospectAfterError = new(bool)
	cmd.Flags().BoolVarP(cmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
}

func SetupIntrospectBeforeError(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.IntrospectBeforeError = new(bool)
	cmd.Flags().BoolVarP(cmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")
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

func SetupSkipBuild(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SkipBuild = new(bool)
	cmd.Flags().BoolVarP(cmdData.SkipBuild, "skip-build", "Z", GetBoolEnvironmentDefaultFalse("WERF_SKIP_BUILD"), "Disable building of docker images, cached images in the repo should exist in the repo if werf.yaml contains at least one image description (default $WERF_SKIP_BUILD)")
}

func SetupStubTags(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.StubTags = new(bool)
	cmd.Flags().BoolVarP(cmdData.StubTags, "stub-tags", "", GetBoolEnvironmentDefaultFalse("WERF_STUB_TAGS"), "Use stubs instead of real tags (default $WERF_STUB_TAGS)")
}

func SetupFollow(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.Follow = new(bool)
	cmd.Flags().BoolVarP(cmdData.Follow, "follow", "", GetBoolEnvironmentDefaultFalse("WERF_FOLLOW"), "Follow git HEAD and run command for each new commit (default $WERF_FOLLOW)")
}

func allStagesNames() []string {
	var stageNames []string
	for _, stageName := range stage.AllStages {
		stageNames = append(stageNames, string(stageName))
	}

	return stageNames
}

func GetBoolEnvironmentDefaultFalse(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func GetBoolEnvironmentDefaultTrue(environmentName string) bool {
	switch os.Getenv(environmentName) {
	case "0", "false", "no":
		return false
	default:
		return true
	}
}

func ConvertInt32Value(v string) (int32, error) {
	res, err := ConvertIntValue(v, 32)
	if err != nil {
		return 0, err
	}
	return int32(res), nil
}

func ConvertIntValue(v string, bitSize int) (int64, error) {
	vInt, err := strconv.ParseInt(v, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("bad integer value '%s': %s", v, err)
	}
	return vInt, nil
}

func getInt64EnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value '%s': %s", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func getIntEnvVar(varName string) (*int64, error) {
	if v := os.Getenv(varName); v != "" {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value '%s': %s", varName, v, err)
		}

		res := new(int64)
		*res = vInt

		return res, nil
	}

	return nil, nil
}

func getUint64EnvVar(varName string) (*uint64, error) {
	if v := os.Getenv(varName); v != "" {
		vUint, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad %s variable value '%s': %s", varName, v, err)
		}

		res := new(uint64)
		*res = vUint

		return res, nil
	}

	return nil, nil
}

func GetParallelTasksLimit(cmdData *CmdData) (int64, error) {
	v, err := getInt64EnvVar("WERF_PARALLEL_TASKS_LIMIT")
	if err != nil {
		return 0, err
	}
	if v == nil {
		v = cmdData.ParallelTasksLimit
	}
	if *v <= 0 {
		return -1, nil
	} else {
		return *v, nil
	}
}

func GetStagesStorageAddress(cmdData *CmdData) (string, error) {
	if *cmdData.StagesStorage == "" || *cmdData.StagesStorage == storage.LocalStorageAddress {
		return "", fmt.Errorf("--repo=ADDRESS param required")
	}

	return *cmdData.StagesStorage, nil
}

func GetOptionalStagesStorageAddress(cmdData *CmdData) string {
	if *cmdData.StagesStorage == "" {
		return storage.LocalStorageAddress
	}

	return *cmdData.StagesStorage
}

func GetStagesStorage(stagesStorageAddress string, containerRuntime container_runtime.ContainerRuntime, cmdData *CmdData) (storage.StagesStorage, error) {
	if err := ValidateRepoImplementation(*cmdData.CommonRepoData.Implementation); err != nil {
		return nil, err
	}

	return storage.NewStagesStorage(
		stagesStorageAddress,
		containerRuntime,
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				Implementation: *cmdData.CommonRepoData.Implementation,
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *cmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
					DockerHubUsername:     *cmdData.CommonRepoData.DockerHubUsername,
					DockerHubPassword:     *cmdData.CommonRepoData.DockerHubPassword,
					DockerHubToken:        *cmdData.CommonRepoData.DockerHubToken,
					GitHubToken:           *cmdData.CommonRepoData.GitHubToken,
					HarborUsername:        *cmdData.CommonRepoData.HarborUsername,
					HarborPassword:        *cmdData.CommonRepoData.HarborPassword,
					QuayToken:             *cmdData.CommonRepoData.QuayToken,
				},
			},
		},
	)
}

func GetSecondaryStagesStorageList(stagesStorage storage.StagesStorage, containerRuntime container_runtime.ContainerRuntime, cmdData *CmdData) ([]storage.StagesStorage, error) {
	var res []storage.StagesStorage
	if stagesStorage.Address() != storage.LocalStorageAddress {
		localStagesStorage, err := storage.NewStagesStorage(storage.LocalStorageAddress, containerRuntime, storage.StagesStorageOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to create local secondary stages storage: %s", err)
		}
		res = append(res, localStagesStorage)
	}

	for _, address := range *cmdData.SecondaryStagesStorage {
		repoStagesStorage, err := storage.NewStagesStorage(address, containerRuntime, storage.StagesStorageOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to create secondary stages storage at %s: %s", address, err)
		}
		res = append(res, repoStagesStorage)
	}

	return res, nil
}

func GetOptionalWerfConfig(ctx context.Context, projectDir string, cmdData *CmdData, localGitRepo *git_repo.Local, opts config.WerfConfigOptions) (*config.WerfConfig, error) {
	werfConfigPath, err := GetWerfConfigPath(projectDir, cmdData, false, localGitRepo, opts)
	if err != nil {
		return nil, err
	}

	if werfConfigPath != "" {
		werfConfigTemplatesDir := GetWerfConfigTemplatesDir(projectDir, cmdData)
		return config.GetWerfConfig(ctx, werfConfigPath, werfConfigTemplatesDir, localGitRepo, opts)
	}

	return nil, nil
}

func GetRequiredWerfConfig(ctx context.Context, projectDir string, cmdData *CmdData, localGitRepo *git_repo.Local, opts config.WerfConfigOptions) (*config.WerfConfig, error) {
	werfConfigPath, err := GetWerfConfigPath(projectDir, cmdData, true, localGitRepo, opts)
	if err != nil {
		return nil, err
	}

	werfConfigTemplatesDir := GetWerfConfigTemplatesDir(projectDir, cmdData)

	return config.GetWerfConfig(ctx, werfConfigPath, werfConfigTemplatesDir, localGitRepo, opts)
}

func GetWerfConfigPath(projectDir string, cmdData *CmdData, required bool, localGitRepo *git_repo.Local, opts config.WerfConfigOptions) (string, error) {
	var configPathToCheck []string

	customConfigPath := *cmdData.ConfigPath
	if customConfigPath != "" {
		configPathToCheck = append(configPathToCheck, customConfigPath)
	} else {
		for _, werfDefaultConfigName := range []string{"werf.yml", "werf.yaml"} {
			configPathToCheck = append(configPathToCheck, filepath.Join(projectDir, werfDefaultConfigName))
		}
	}

	for _, werfConfigPath := range configPathToCheck {
		if opts.DisableDeterminism || localGitRepo == nil {
			if exists, err := util.FileExists(werfConfigPath); err != nil {
				return "", err
			} else if exists {
				return werfConfigPath, nil
			}
		} else {
			ctx := BackgroundContext()
			commit, err := localGitRepo.HeadCommit(ctx)
			if err != nil {
				return "", fmt.Errorf("unable to get local repo head commit: %s", err)
			}

			relPath := util.GetRelativeToBaseFilepath(projectDir, werfConfigPath)
			if exists, err := localGitRepo.IsFileExists(commit, relPath); err != nil {
				return "", fmt.Errorf("unable to check %s existance in the local git repo: %s", relPath, err)
			} else if exists {
				_, isDataIndentical, err := git_repo.GetFileDataFromGitAndCompareWithLocal(localGitRepo, commit, projectDir, relPath)
				if err != nil {
					return "", fmt.Errorf("unable to compare local git repo and workdir file %s data: %s", relPath, err)
				}

				if !isDataIndentical {
					logboek.Context(ctx).Warn().LogF("WARNING: In deterministic mode uncommitted file %s was not taken into account\n", relPath)
				}

				return relPath, nil
			}
		}
	}

	if required {
		if opts.DisableDeterminism || localGitRepo == nil {
			return "", fmt.Errorf("werf configuration file not found (%s)", strings.Join(configPathToCheck, ", "))
		} else {
			return "", fmt.Errorf("werf configuration file not found (%s) in the local git repo", strings.Join(configPathToCheck, ", "))
		}
	}

	return "", nil
}

func GetWerfConfigTemplatesDir(projectDir string, cmdData *CmdData) string {
	customConfigTemplatesDir := *cmdData.ConfigTemplatesDir
	if customConfigTemplatesDir != "" {
		return util.GetRelativeToBaseFilepath(projectDir, customConfigTemplatesDir)
	} else {
		return util.GetRelativeToBaseFilepath(projectDir, filepath.Join(projectDir, ".werf"))
	}
}

func GetProjectDir(cmdData *CmdData) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if *cmdData.Dir != "" {
		if filepath.IsAbs(*cmdData.Dir) {
			return *cmdData.Dir, nil
		} else {
			return filepath.Clean(filepath.Join(currentDir, *cmdData.Dir)), nil
		}
	}

	return currentDir, nil
}

func GetHelmChartDir(projectDir string, cmdData *CmdData, werfConfig *config.WerfConfig) (string, error) {
	var helmChartDir string

	if werfConfig.Meta.Deploy.HelmChartDir != nil && *werfConfig.Meta.Deploy.HelmChartDir != "" {
		helmChartDir = *werfConfig.Meta.Deploy.HelmChartDir
	} else {
		helmChartDir = ".helm"
	}

	return helmChartDir, nil
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
	for _, imageAndStage := range *cmdData.StagesToIntrospect {
		var imageName, stageName string

		parts := strings.SplitN(imageAndStage, "/", 2)
		if len(parts) == 1 {
			imageName = "*"
			stageName = parts[0]
		} else {
			if parts[0] != "~" {
				imageName = parts[0]
			}

			stageName = parts[1]
		}

		if imageName != "*" && !werfConfig.HasImageOrArtifact(imageName) {
			return introspectOptions, fmt.Errorf("specified image %s (%s) is not defined in werf.yaml", imageName, imageAndStage)
		}

		if !isStageExist(stageName) {
			return introspectOptions, fmt.Errorf("specified stage name %s (%s) is not exist", stageName, imageAndStage)
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

	if *cmdData.LogQuiet {
		logboek.SetAcceptedLevel(level.Error)
		logboek.Streams().Mute()
	} else if *cmdData.LogDebug {
		logboek.SetAcceptedLevel(level.Debug)
		logboek.Streams().EnablePrefixWithTime()
		logboek.Streams().SetPrefixStyle(style.Details())
	} else if *cmdData.LogVerbose {
		logboek.SetAcceptedLevel(level.Info)
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

func ProcessLogOptionsDefaultQuiet(cmdData *CmdData) error {
	if !*cmdData.LogQuiet {
		logboek.Streams().Unmute()
		logboek.SetAcceptedLevel(level.Default)
	}

	if err := ProcessLogColorMode(cmdData); err != nil {
		return err
	}

	if *cmdData.LogDebug {
		logboek.Streams().Unmute()
		logboek.SetAcceptedLevel(level.Debug)
		logboek.Streams().EnablePrefixWithTime()
		logboek.Streams().SetPrefixStyle(style.Details())
	} else if *cmdData.LogVerbose {
		logboek.Streams().Unmute()
		logboek.SetAcceptedLevel(level.Info)
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

func ProcessLogColorMode(cmdData *CmdData) error {
	logColorMode := *cmdData.LogColorMode

	switch logColorMode {
	case "auto":
	case "on":
		logboek.Streams().EnableStyle()
	case "off":
		logboek.Streams().DisableStyle()
	default:
		return fmt.Errorf("bad log color mode '%s': on, off and auto modes are supported", logColorMode)
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
		pInt64, err := getInt64EnvVar("WERF_LOG_TERMINAL_WIDTH")
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

func DockerRegistryInit(cmdData *CmdData) error {
	return docker_registry.Init(BackgroundContext(), *cmdData.InsecureRegistry, *cmdData.SkipTlsVerifyRegistry)
}

func ValidateRepoImplementation(implementation string) error {
	supportedValues := docker_registry.ImplementationList()
	supportedValues = append(supportedValues, "auto", "")

	for _, supportedImplementation := range supportedValues {
		if supportedImplementation == implementation {
			return nil
		}
	}

	return fmt.Errorf("specified docker registry implementation '%s' is not supported", implementation)
}

func ValidateMinimumNArgs(minArgs int, args []string, cmd *cobra.Command) error {
	if len(args) < minArgs {
		PrintHelp(cmd)
		return fmt.Errorf("requires at least %d arg(s), received %d", minArgs, len(args))
	}

	return nil
}

func ValidateMaximumNArgs(maxArgs int, args []string, cmd *cobra.Command) error {
	if len(args) > maxArgs {
		PrintHelp(cmd)
		return fmt.Errorf("accepts at most %d arg(s), received %d", maxArgs, len(args))
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

	logboek.Default().LogFHighlight("Running time %0.2f seconds\n", time.Now().Sub(t).Seconds())

	return err
}

func LogVersion() {
	logboek.LogF("Version: %s\n", werf.Version)
}

func TerminateWithError(errMsg string, exitCode int) {
	msg := fmt.Sprintf("Error: %s", errMsg)
	msg = strings.TrimSuffix(msg, "\n")

	logboek.Streams().DisableLineWrapping()
	logboek.Error().LogLn(msg)
	os.Exit(exitCode)
}

func SetupVirtualMerge(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMerge = new(bool)
	cmd.Flags().BoolVarP(cmdData.VirtualMerge, "virtual-merge", "", GetBoolEnvironmentDefaultFalse("WERF_VIRTUAL_MERGE"), "Enable virtual/ephemeral merge commit mode when building current application state ($WERF_VIRTUAL_MERGE by default)")
}

func SetupVirtualMergeFromCommit(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMergeFromCommit = new(string)
	cmd.Flags().StringVarP(cmdData.VirtualMergeFromCommit, "virtual-merge-from-commit", "", os.Getenv("WERF_VIRTUAL_MERGE_FROM_COMMIT"), "Commit hash for virtual/ephemeral merge commit with new changes introduced in the pull request ($WERF_VIRTUAL_MERGE_FROM_COMMIT by default)")
}

func SetupVirtualMergeIntoCommit(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.VirtualMergeIntoCommit = new(string)
	cmd.Flags().StringVarP(cmdData.VirtualMergeIntoCommit, "virtual-merge-into-commit", "", os.Getenv("WERF_VIRTUAL_MERGE_INTO_COMMIT"), "Commit hash for virtual/ephemeral merge commit which is base for changes introduced in the pull request ($WERF_VIRTUAL_MERGE_INTO_COMMIT by default)")
}

func BackgroundContext() context.Context {
	return logboek.NewContext(context.Background(), logboek.DefaultLogger())
}
