package ci_env

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

var cmdData struct {
	TaggingStrategy string
	AsFile          bool
	Shell           string
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "ci-env CI_SYSTEM",
		DisableFlagsInUseLine: true,
		Short:                 "Generate werf environment variables for specified CI system",
		Long: `Generate werf environment variables for specified CI system.

Currently supported only GitLab CI`,
		Example: `  # Load generated werf environment variables on GitLab job runner
  $ . $(werf ci-env gitlab --as-file)

  # Load generated werf environment variables on GitLab job runner using powershell
  $ Invoke-Expression -Command "werf ci-env gitlab --as-file --shell powershell" | Out-String -OutVariable WERF_CI_ENV_SCRIPT_PATH
  $ . $WERF_CI_ENV_SCRIPT_PATH.Trim()

  # Load generated werf environment variables on GitLab job runner using cmd.exe
  $ FOR /F "tokens=*" %g IN ('werf ci-env gitlab --as-file --shell cmdexe') do (SET WERF_CI_ENV_SCRIPT_PATH=%g)
  $ %WERF_CI_ENV_SCRIPT_PATH%`,
		RunE: runCIEnv,
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command will copy specified or default (~/.docker) config to the temporary directory and may perform additional login with new config.")

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.TaggingStrategy, "tagging-strategy", "", "stages-signature", `* stages-signature: always use '--tag-by-stages-signature' option to tag all published images by corresponding stages-signature;
* tag-or-branch: generate auto '--tag-git-branch' or '--tag-git-tag' tag by specified CI_SYSTEM environment variables.`)
	cmd.Flags().BoolVarP(&cmdData.AsFile, "as-file", "", common.GetBoolEnvironmentDefaultFalse("WERF_AS_FILE"), "Create the script and print the path for sourcing (default $WERF_AS_FILE).")
	cmd.Flags().StringVarP(&cmdData.Shell, "shell", "", "WERF_SHELL", "Set to cmdexe, powershell or use the default behaviour that is compatible with any unix shell (default $WERF_SHELL).")

	return cmd
}

func runCIEnv(cmd *cobra.Command, args []string) error {
	if err := common.ProcessLogOptions(&commonCmdData); err != nil {
		common.PrintHelp(cmd)
		return err
	}

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
		return err
	}

	switch cmdData.Shell {
	case "", "default", "cmdexe", "powershell":
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided shell '%s' not supported", cmdData.Shell)
	}

	switch cmdData.TaggingStrategy {
	case "tag-or-branch", "stages-signature":
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided tagging-strategy '%s' not supported", cmdData.TaggingStrategy)
	}

	var w io.Writer
	if cmdData.AsFile {
		w = bytes.NewBuffer(nil)
	} else {
		w = os.Stdout
	}

	ciSystem := args[0]
	switch ciSystem {
	case "gitlab":
		err := generateGitlabEnvs(w, cmdData.TaggingStrategy)
		if err != nil {
			if !cmdData.AsFile {
				writeError(w, err.Error())
			}
			return err
		}
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided ci system '%s' not supported", ciSystem)
	}

	if cmdData.AsFile {
		sourceFilePath, err := createSourceFile(w.(*bytes.Buffer).Bytes())
		if err != nil {
			return err
		}

		fmt.Println(sourceFilePath)
	}

	return nil
}

func generateGitlabEnvs(w io.Writer, taggingStrategy string) error {
	dockerConfigPath := *commonCmdData.DockerConfig
	if *commonCmdData.DockerConfig == "" {
		dockerConfigPath = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	tmp_manager.AutoGCEnabled = false

	dockerConfig, err := tmp_manager.CreateDockerConfigDir(dockerConfigPath)
	if err != nil {
		return fmt.Errorf("unable to create tmp docker config: %s", err)
	}

	// Init with new docker config dir
	if err := docker.Init(dockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	ciRegistryImage := os.Getenv("CI_REGISTRY_IMAGE")
	ciJobToken := os.Getenv("CI_JOB_TOKEN")

	var imagesUsername, imagesPassword string
	var imagesRepoImplementation string
	var doLogin bool
	if ciRegistryImage != "" && ciJobToken != "" {
		imagesUsername = "gitlab-ci-token"
		imagesPassword = ciJobToken
		doLogin = true

		imagesRepoImplementation = docker_registry.GitLabRegistryImplementationName
	}

	if doLogin {
		err := docker.Login(imagesUsername, imagesPassword, ciRegistryImage)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %s", ciRegistryImage, err)
		}
	}

	writeHeader(w, "DOCKER CONFIG", false)
	writeExportCommand(w, "DOCKER_CONFIG", dockerConfig, true)

	writeHeader(w, "STAGES_STORAGE", true)
	writeExportCommand(w, "WERF_STAGES_STORAGE", fmt.Sprintf("%s/stages", ciRegistryImage), false)

	writeHeader(w, "IMAGES REPO", true)
	writeExportCommand(w, "WERF_IMAGES_REPO", ciRegistryImage, false)
	writeExportCommand(w, "WERF_IMAGES_REPO_IMPLEMENTATION", imagesRepoImplementation, false)

	writeHeader(w, "TAGGING", true)
	switch taggingStrategy {
	case "tag-or-branch":
		var ciGitTag, ciGitBranch string

		if os.Getenv("CI_BUILD_TAG") != "" {
			ciGitTag = os.Getenv("CI_BUILD_TAG")
		} else if os.Getenv("CI_COMMIT_TAG") != "" {
			ciGitTag = os.Getenv("CI_COMMIT_TAG")
		} else if os.Getenv("CI_BUILD_REF_NAME") != "" {
			ciGitBranch = os.Getenv("CI_BUILD_REF_NAME")
		} else if os.Getenv("CI_COMMIT_REF_NAME") != "" {
			ciGitBranch = os.Getenv("CI_COMMIT_REF_NAME")
		}

		if ciGitTag != "" {
			writeExportCommand(w, "WERF_TAG_GIT_TAG", slug.DockerTag(ciGitTag), false)
		}
		if ciGitBranch != "" {
			writeExportCommand(w, "WERF_TAG_GIT_BRANCH", slug.DockerTag(ciGitBranch), false)
		}

		if ciGitTag == "" && ciGitBranch == "" {
			return fmt.Errorf("none of enviroment variables $WERF_TAG_GIT_TAG=$CI_COMMIT_TAG or $WERF_TAG_GIT_BRANCH=$CI_COMMIT_REF_NAME for '%s' strategy are detected", cmdData.TaggingStrategy)
		}
	case "stages-signature":
		writeExportCommand(w, "WERF_TAG_BY_STAGES_SIGNATURE", "true", false)
	}

	writeHeader(w, "DEPLOY", true)
	writeExportCommand(w, "WERF_ENV", os.Getenv("CI_ENVIRONMENT_SLUG"), false)

	var projectGit string
	ciProjectUrlEnv := os.Getenv("CI_PROJECT_URL")
	if ciProjectUrlEnv != "" {
		projectGit = fmt.Sprintf("project.werf.io/git=%s", ciProjectUrlEnv)
	}
	writeExportCommand(w, "WERF_ADD_ANNOTATION_PROJECT_GIT", projectGit, false)

	var ciCommit string
	ciCommitShaEnv := os.Getenv("CI_COMMIT_SHA")
	if ciCommitShaEnv != "" {
		ciCommit = fmt.Sprintf("ci.werf.io/commit=%s", ciCommitShaEnv)
	}
	writeExportCommand(w, "WERF_ADD_ANNOTATION_CI_COMMIT", ciCommit, false)

	var gitlabCIPipelineUrl string
	ciPipelineIdEnv := os.Getenv("CI_PIPELINE_ID")
	if ciProjectUrlEnv != "" && ciPipelineIdEnv != "" {
		gitlabCIPipelineUrl = fmt.Sprintf("gitlab.ci.werf.io/pipeline-url=%s/pipelines/%s", ciProjectUrlEnv, ciPipelineIdEnv)
	}
	writeExportCommand(w, "WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL", gitlabCIPipelineUrl, false)

	var gitlabCiJobUrl string
	ciJobIdEnv := os.Getenv("CI_JOB_ID")
	if ciProjectUrlEnv != "" && os.Getenv("CI_JOB_ID") != "" {
		gitlabCiJobUrl = fmt.Sprintf("gitlab.ci.werf.io/job-url=%s/-/jobs/%s", ciProjectUrlEnv, ciJobIdEnv)
	}
	writeExportCommand(w, "WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL", gitlabCiJobUrl, false)

	cleanupConfig, err := getCleanupConfig()
	if err != nil {
		return fmt.Errorf("unable to get cleanup config: %s", err)
	}

	writeHeader(w, "IMAGE CLEANUP POLICIES", true)
	writeExportCommand(w, "WERF_GIT_TAG_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyLimit), false)
	writeExportCommand(w, "WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyExpiryDays), false)
	writeExportCommand(w, "WERF_GIT_COMMIT_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyLimit), false)
	writeExportCommand(w, "WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyExpiryDays), false)
	writeExportCommand(w, "WERF_STAGES_SIGNATURE_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.StagesSignatureStrategyLimit), false)
	writeExportCommand(w, "WERF_STAGES_SIGNATURE_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.StagesSignatureStrategyExpiryDays), false)

	writeHeader(w, "OTHER", true)

	werfLogColorMode := "on"
	ciServerVersion := os.Getenv("CI_SERVER_VERSION")
	if ciServerVersion != "" {
		currentVersion, err := semver.NewVersion(ciServerVersion)
		if err == nil {
			colorWorkTillVersion, _ := semver.NewVersion("12.1.3")
			colorWorkSinceVersion, _ := semver.NewVersion("12.2.0")

			if currentVersion.GreaterThan(colorWorkTillVersion) && currentVersion.LessThan(colorWorkSinceVersion) {
				werfLogColorMode = "off"
			}
		}
	}

	writeExportCommand(w, "WERF_LOG_COLOR_MODE", werfLogColorMode, false)
	writeExportCommand(w, "WERF_LOG_PROJECT_DIR", "1", false)
	writeExportCommand(w, "WERF_ENABLE_PROCESS_EXTERMINATOR", "1", false)
	writeExportCommand(w, "WERF_LOG_TERMINAL_WIDTH", "95", false)

	return nil
}

func writeError(w io.Writer, errMsg string) {
	if *commonCmdData.LogVerbose {
		_, _ = fmt.Fprintln(w, "echo")
		_, _ = fmt.Fprintf(w, "echo 'Error: %s'\n", errMsg)
	}

	_, _ = fmt.Fprintf(w, "exit 1\n")
	_, _ = fmt.Fprintln(w)
}

func writeHeader(w io.Writer, header string, withNewLine bool) {
	var commentSigns string
	switch cmdData.Shell {
	case "cmdexe":
		commentSigns = "::"
	default:
		commentSigns = "###"
	}

	header = fmt.Sprintf("%s %s", commentSigns, header)

	if withNewLine {
		_, _ = fmt.Fprintln(w)
	}
	_, _ = fmt.Fprintln(w, header)

	if *commonCmdData.LogVerbose {
		if withNewLine {
			_, _ = fmt.Fprintln(w, "echo")
		}
		echoHeader := fmt.Sprintf("echo '%s'", header)
		_, _ = fmt.Fprintln(w, echoHeader)
	}
}

func writeExportCommand(w io.Writer, key, value string, override bool) {
	var commentSign string
	switch cmdData.Shell {
	case "cmdexe":
		commentSign = "::"
	default:
		commentSign = "#"
	}

	if !override && os.Getenv(key) != "" {
		skipComment := fmt.Sprintf("%s skip %s=\"%s\"", commentSign, key, os.Getenv(key))
		_, _ = fmt.Fprintln(w, skipComment)

		if *commonCmdData.LogVerbose {
			echoSkip := fmt.Sprintf("echo '%s'", skipComment)
			_, _ = fmt.Fprintln(w, echoSkip)
		}

		return
	}

	var exportFormat string
	switch cmdData.Shell {
	case "powershell":
		exportFormat = "$Env:%s = \"%s\""
	case "cmd.exe":
		exportFormat = "set %s=%s"
	default:
		exportFormat = "export %s=\"%s\""
	}

	exportCommand := fmt.Sprintf(exportFormat, key, value)
	if value == "" {
		exportCommand = fmt.Sprintf("%s %s", commentSign, exportCommand)
	}

	_, _ = fmt.Fprintln(w, exportCommand)

	if *commonCmdData.LogVerbose {
		echoExportCommand := fmt.Sprintf("echo '%s'", exportCommand)
		_, _ = fmt.Fprintln(w, echoExportCommand)
	}
}

type CleanupConfig struct {
	GitTagStrategyLimit               int `yaml:"gitTagStrategyLimit"`
	GitTagStrategyExpiryDays          int `yaml:"gitTagStrategyExpiryDays"`
	GitCommitStrategyLimit            int `yaml:"gitCommitStrategyLimit"`
	GitCommitStrategyExpiryDays       int `yaml:"gitCommitStrategyExpiryDays"`
	StagesSignatureStrategyExpiryDays int `yaml:"stagesSignatureStrategyExpiryDays"`
	StagesSignatureStrategyLimit      int `yaml:"stagesSignatureStrategyLimit"`
}

func getCleanupConfig() (CleanupConfig, error) {
	configPath := filepath.Join(werf.GetHomeDir(), "config", "cleanup.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return CleanupConfig{
			GitTagStrategyLimit:               10,
			GitTagStrategyExpiryDays:          30,
			GitCommitStrategyLimit:            50,
			GitCommitStrategyExpiryDays:       30,
			StagesSignatureStrategyLimit:      -1,
			StagesSignatureStrategyExpiryDays: -1,
		}, nil
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return CleanupConfig{}, fmt.Errorf("error reading %s: %s", configPath, err)
	}

	config := CleanupConfig{}
	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		return CleanupConfig{}, fmt.Errorf("bad config yaml %s: %s", configPath, err)
	}

	return config, nil
}

func createSourceFile(data []byte) (string, error) {
	sourceDir := filepath.Join(werf.GetServiceDir(), "tmp", "ci_env")
	err := os.MkdirAll(sourceDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return "", err
	}

	// keep no more than 100 source files, remain 50 during cleaning
	if len(files) >= 100 {
		for i := len(files) - 50; i >= 0; i-- {
			file := files[i]
			if file.IsDir() {
				continue
			}

			sourceFilePath := filepath.Join(sourceDir, file.Name())
			if err := os.Remove(sourceFilePath); err != nil {
				return "", err
			}
		}
	}

	tempFilePattern := fmt.Sprintf("source_%d_*", time.Now().Unix())
	switch cmdData.Shell {
	case "cmdexe":
		tempFilePattern += ".bat"
	case "powershell":
		tempFilePattern += ".ps1"
	}

	f, err := ioutil.TempFile(sourceDir, tempFilePattern)
	if err != nil {
		return "", err
	}

	if _, err := f.Write(data); err != nil {
		return "", err
	}

	return f.Name(), nil
}
