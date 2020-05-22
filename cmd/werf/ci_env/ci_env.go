package ci_env

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

var cmdData struct {
	TaggingStrategy string
	AsFile          bool
	AsEnvFile       bool
	OutputFilePath  string
	Shell           string
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "ci-env CI_SYSTEM",
		DisableFlagsInUseLine: true,
		Short:                 "Generate werf environment variables for specified CI system",
		Long: `Generate werf environment variables for specified CI system.

Currently supported only GitLab (gitlab) and GitHub (github) CI systems`,
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

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command will copy specified or default (~/.docker) config to the temporary directory and may perform additional login with new config.")

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.TaggingStrategy, "tagging-strategy", "", "stages-signature", `* stages-signature: always use '--tag-by-stages-signature' option to tag all published images by corresponding stages-signature;
* tag-or-branch: generate auto '--tag-git-branch' or '--tag-git-tag' tag by specified CI_SYSTEM environment variables.`)
	cmd.Flags().BoolVarP(&cmdData.AsFile, "as-file", "", common.GetBoolEnvironmentDefaultFalse("WERF_AS_FILE"), "Create the script and print the path for sourcing (default $WERF_AS_FILE).")
	cmd.Flags().BoolVarP(&cmdData.AsEnvFile, "as-env-file", "", common.GetBoolEnvironmentDefaultFalse("WERF_AS_ENV_FILE"), "Create the .env file and print the path for sourcing (default $WERF_AS_ENV_FILE).")
	cmd.Flags().StringVarP(&cmdData.OutputFilePath, "output-file-path", "o", os.Getenv("WERF_OUTPUT_FILE_PATH"), "Write to custom file (default $WERF_OUTPUT_FILE_PATH).")
	cmd.Flags().StringVarP(&cmdData.Shell, "shell", "", os.Getenv("WERF_SHELL"), "Set to cmdexe, powershell or use the default behaviour that is compatible with any unix shell (default $WERF_SHELL).")

	return cmd
}

func runCIEnv(cmd *cobra.Command, args []string) error {
	//if err := common.ProcessLogOptions(&commonCmdData); err != nil {
	//	common.PrintHelp(cmd)
	//	return err
	//}
	logging.EnableLogQuiet()

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
	if cmdData.AsFile || cmdData.AsEnvFile {
		w = bytes.NewBuffer(nil)
	} else {
		w = os.Stdout
	}

	ciSystem := args[0]
	switch ciSystem {
	case "github":
		err := generateGithubEnvs(w, cmdData.TaggingStrategy)
		if err != nil {
			if !cmdData.AsFile && !cmdData.AsEnvFile {
				writeError(w, err.Error())
			}
			return err
		}
	case "gitlab":
		err := generateGitlabEnvs(w, cmdData.TaggingStrategy)
		if err != nil {
			if !cmdData.AsFile && !cmdData.AsEnvFile {
				writeError(w, err.Error())
			}
			return err
		}
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided ci system '%s' not supported", ciSystem)
	}

	if cmdData.AsFile || cmdData.AsEnvFile {
		sourceFilePath, err := createSourceFile(w.(*bytes.Buffer).Bytes())
		if err != nil {
			return err
		}

		if cmdData.OutputFilePath == "" {
			fmt.Println(sourceFilePath)
		}
	}

	return nil
}

func generateGitlabEnvs(w io.Writer, taggingStrategy string) error {
	dockerConfig, err := generateSessionDockerConfigDir()
	if err != nil {
		return err
	}

	ciRegistryImageEnv := os.Getenv("CI_REGISTRY_IMAGE")
	ciJobTokenEnv := os.Getenv("CI_JOB_TOKEN")

	var imagesUsername, imagesPassword string
	var doLogin bool
	if ciRegistryImageEnv != "" && ciJobTokenEnv != "" {
		imagesUsername = "gitlab-ci-token"
		imagesPassword = ciJobTokenEnv
		doLogin = true
	}

	var stagesStorageRepoImplementation string
	var imagesRepoImplementation string

	ciRegistryEnv := os.Getenv("CI_REGISTRY")
	werfStagesStorageEnv := os.Getenv("WERF_STAGES_STORAGE")
	werfImagesRepoEnv := os.Getenv("WERF_IMAGES_REPO")

	if werfStagesStorageEnv == "" || strings.HasPrefix(werfStagesStorageEnv, ciRegistryEnv) {
		stagesStorageRepoImplementation = docker_registry.GitLabRegistryImplementationName
	}

	if werfImagesRepoEnv == "" || strings.HasPrefix(werfStagesStorageEnv, ciRegistryEnv) {
		imagesRepoImplementation = docker_registry.GitLabRegistryImplementationName
	}

	if doLogin {
		err := docker.Login(imagesUsername, imagesPassword, ciRegistryImageEnv)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %s", ciRegistryImageEnv, err)
		}
	}

	writeHeader(w, "DOCKER CONFIG", false)
	writeEnv(w, "DOCKER_CONFIG", dockerConfig, true)

	writeHeader(w, "STAGES_STORAGE", true)
	writeEnv(w, "WERF_STAGES_STORAGE", fmt.Sprintf("%s/stages", ciRegistryImageEnv), false)
	if stagesStorageRepoImplementation != "" {
		writeEnv(w, "WERF_STAGES_STORAGE_REPO_IMPLEMENTATION", stagesStorageRepoImplementation, false)
	}

	writeHeader(w, "IMAGES REPO", true)
	writeEnv(w, "WERF_IMAGES_REPO", ciRegistryImageEnv, false)
	if imagesRepoImplementation != "" {
		writeEnv(w, "WERF_IMAGES_REPO_IMPLEMENTATION", imagesRepoImplementation, false)
	}

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
			writeEnv(w, "WERF_TAG_GIT_TAG", slug.DockerTag(ciGitTag), false)
		}
		if ciGitBranch != "" {
			writeEnv(w, "WERF_TAG_GIT_BRANCH", slug.DockerTag(ciGitBranch), false)
		}

		if ciGitTag == "" && ciGitBranch == "" {
			return fmt.Errorf("none of enviroment variables $WERF_TAG_GIT_TAG=$CI_COMMIT_TAG or $WERF_TAG_GIT_BRANCH=$CI_COMMIT_REF_NAME for '%s' strategy are detected", cmdData.TaggingStrategy)
		}
	case "stages-signature":
		writeEnv(w, "WERF_TAG_BY_STAGES_SIGNATURE", "true", false)
	}

	writeHeader(w, "DEPLOY", true)
	writeEnv(w, "WERF_ENV", os.Getenv("CI_ENVIRONMENT_SLUG"), false)

	var projectGit string
	ciProjectUrlEnv := os.Getenv("CI_PROJECT_URL")
	if ciProjectUrlEnv != "" {
		projectGit = fmt.Sprintf("project.werf.io/git=%s", ciProjectUrlEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_PROJECT_GIT", projectGit, false)

	var ciCommit string
	ciCommitShaEnv := os.Getenv("CI_COMMIT_SHA")
	if ciCommitShaEnv != "" {
		ciCommit = fmt.Sprintf("ci.werf.io/commit=%s", ciCommitShaEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_CI_COMMIT", ciCommit, false)

	var gitlabCIPipelineUrl string
	ciPipelineIdEnv := os.Getenv("CI_PIPELINE_ID")
	if ciProjectUrlEnv != "" && ciPipelineIdEnv != "" {
		gitlabCIPipelineUrl = fmt.Sprintf("gitlab.ci.werf.io/pipeline-url=%s/pipelines/%s", ciProjectUrlEnv, ciPipelineIdEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL", gitlabCIPipelineUrl, false)

	var gitlabCiJobUrl string
	ciJobIdEnv := os.Getenv("CI_JOB_ID")
	if ciProjectUrlEnv != "" && os.Getenv("CI_JOB_ID") != "" {
		gitlabCiJobUrl = fmt.Sprintf("gitlab.ci.werf.io/job-url=%s/-/jobs/%s", ciProjectUrlEnv, ciJobIdEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL", gitlabCiJobUrl, false)

	if err = generateImageCleanupPolicies(w); err != nil {
		return err
	}

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

	writeEnv(w, "WERF_LOG_COLOR_MODE", werfLogColorMode, false)
	writeEnv(w, "WERF_LOG_PROJECT_DIR", "1", false)
	writeEnv(w, "WERF_ENABLE_PROCESS_EXTERMINATOR", "1", false)
	writeEnv(w, "WERF_LOG_TERMINAL_WIDTH", "95", false)

	return nil
}

func generateGithubEnvs(w io.Writer, taggingStrategy string) error {
	dockerConfigDir, err := generateSessionDockerConfigDir()
	if err != nil {
		return err
	}

	githubRegistry := "docker.pkg.github.com"
	ciGithubToken := os.Getenv("GITHUB_TOKEN")
	ciGithubActor := os.Getenv("GITHUB_ACTOR")
	if ciGithubActor != "" && ciGithubToken != "" {
		err := docker.Login(ciGithubActor, ciGithubToken, githubRegistry)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %s", githubRegistry, err)
		}
	}

	ciGithubOwnerWithProject := os.Getenv("GITHUB_REPOSITORY")

	var imagesRepo, stagesStorageRepo string
	if ciGithubOwnerWithProject != "" {
		projectDir, err := common.GetProjectDir(&commonCmdData)
		if err != nil {
			return fmt.Errorf("getting project dir failed: %s", err)
		}

		werfConfig, err := common.GetOptionalWerfConfig(projectDir, &commonCmdData, true)
		if err != nil {
			return fmt.Errorf("unable to load werf config: %s", err)
		}

		projectRepo := fmt.Sprintf("%s/%s", githubRegistry, ciGithubOwnerWithProject)
		multirepo := projectRepo
		monorepo := fmt.Sprintf("%s/%s", projectRepo, werfConfig.Meta.Project)

		if werfConfig != nil {
			if werfConfig.HasNamelessImage() {
				imagesRepo = monorepo
			} else {
				imagesRepo = multirepo
			}
		} else {
			imagesRepo = monorepo
		}

		stagesStorageRepo = fmt.Sprintf("%s/stages", projectRepo)
	} else if os.Getenv("IMAGES REPO") != "" && os.Getenv("STAGES_STORAGE") == "" {
		stagesStorageRepo = fmt.Sprintf("%s/stages", os.Getenv("IMAGES REPO"))
	}

	writeHeader(w, "DOCKER CONFIG", false)
	writeEnv(w, "DOCKER_CONFIG", dockerConfigDir, true)

	writeHeader(w, "STAGES_STORAGE", true)
	writeEnv(w, "WERF_STAGES_STORAGE", stagesStorageRepo, false)

	writeHeader(w, "IMAGES REPO", true)
	writeEnv(w, "WERF_IMAGES_REPO", imagesRepo, false)

	writeHeader(w, "TAGGING", true)
	switch taggingStrategy {
	case "stages-signature":
		writeEnv(w, "WERF_TAG_BY_STAGES_SIGNATURE", "true", false)
	default:
		return fmt.Errorf("provided tagging-strategy '%s' not supported", taggingStrategy)
	}

	writeHeader(w, "DEPLOY", true)
	var projectGit string
	if ciGithubOwnerWithProject != "" {
		projectGit = fmt.Sprintf("project.werf.io/git=%s", fmt.Sprintf("https://github.com/%s", ciGithubOwnerWithProject))
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_PROJECT_GIT", projectGit, false)

	writeHeader(w, "CLEANUP", true)
	writeEnv(w, "WERF_REPO_GITHUB_TOKEN", ciGithubToken, false)

	var ciCommit string
	ciCommitShaEnv := os.Getenv("GITHUB_SHA")
	if ciCommitShaEnv != "" {
		ciCommit = fmt.Sprintf("ci.werf.io/commit=%s", ciCommitShaEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_CI_COMMIT", ciCommit, false)

	var workflowUrl string
	ciWorkflowRunIdEnv := os.Getenv("GITHUB_RUN_ID")
	if ciGithubOwnerWithProject != "" && ciWorkflowRunIdEnv != "" {
		workflowUrl = fmt.Sprintf("project.werf.io/git=%s", fmt.Sprintf("https://github.com/%s/actions/runs/%s", ciGithubOwnerWithProject, ciWorkflowRunIdEnv))
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_GITHUB_CI_WORKFLOW_URL", workflowUrl, false)

	if err = generateImageCleanupPolicies(w); err != nil {
		return err
	}

	if err = generateOther(w); err != nil {
		return err
	}

	return nil
}

func generateSessionDockerConfigDir() (string, error) {
	dockerConfigPath := *commonCmdData.DockerConfig
	if *commonCmdData.DockerConfig == "" {
		dockerConfigPath = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	tmp_manager.AutoGCEnabled = false

	dockerConfigDir, err := tmp_manager.CreateDockerConfigDir(dockerConfigPath)
	if err != nil {
		return "", fmt.Errorf("unable to create tmp docker config: %s", err)
	}

	// Init with new docker config dir
	if err := docker.Init(dockerConfigDir, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return "", err
	}

	return dockerConfigDir, nil
}

func generateImageCleanupPolicies(w io.Writer) error {
	cleanupConfig, err := getCleanupConfig()
	if err != nil {
		return fmt.Errorf("unable to get cleanup config: %s", err)
	}

	writeHeader(w, "IMAGE CLEANUP POLICIES", true)
	writeEnv(w, "WERF_GIT_TAG_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyLimit), false)
	writeEnv(w, "WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyExpiryDays), false)
	writeEnv(w, "WERF_GIT_COMMIT_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyLimit), false)
	writeEnv(w, "WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyExpiryDays), false)
	writeEnv(w, "WERF_STAGES_SIGNATURE_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.StagesSignatureStrategyLimit), false)
	writeEnv(w, "WERF_STAGES_SIGNATURE_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.StagesSignatureStrategyExpiryDays), false)

	return nil
}

func generateOther(w io.Writer) error {
	writeHeader(w, "OTHER", true)
	writeEnv(w, "WERF_LOG_COLOR_MODE", "on", false)
	writeEnv(w, "WERF_LOG_PROJECT_DIR", "1", false)
	writeEnv(w, "WERF_ENABLE_PROCESS_EXTERMINATOR", "1", false)
	writeEnv(w, "WERF_LOG_TERMINAL_WIDTH", "95", false)

	return nil
}

func writeError(w io.Writer, errMsg string) {
	if *commonCmdData.LogVerbose {
		writeLn(w, echoLine(""))
		writeLn(w, echoLine("Error: "+errMsg))
	}

	writeLn(w, "exit 1")
}

func writeHeader(w io.Writer, header string, withNewLine bool) {
	if withNewLine {
		writeLn(w, "")
	}

	headerLine := commentLine(header)
	writeLn(w, headerLine)

	if *commonCmdData.LogVerbose {
		if withNewLine {
			writeLn(w, echoLine(""))
		}

		echoHeader := echoLine(headerLine)
		writeLn(w, echoHeader)
	}
}

func writeEnv(w io.Writer, key, value string, override bool) {
	envLine := envLine(key, value)

	if !override && os.Getenv(key) != "" {
		skipLine := skipLine(fmt.Sprintf("%s (%s)", envLine, os.Getenv(key)))
		writeLn(w, skipLine)

		if *commonCmdData.LogVerbose {
			writeLn(w, echoLine(skipLine))
		}

		return
	}

	if value == "" {
		envLine = commentLine(envLine)
	}

	writeLn(w, envLine)

	if *commonCmdData.LogVerbose {
		writeLn(w, echoLine(envLine))
	}
}

func writeLn(w io.Writer, message string) {
	_, err := fmt.Fprintln(w, message)
	if err != nil {
		panic("unexpected error: " + err.Error())
	}
}

func envLine(envKey, envValue string) string {
	if cmdData.AsEnvFile {
		return strings.Join([]string{envKey, envValue}, "=")
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

	return fmt.Sprintf(exportFormat, envKey, envValue)
}

func echoLine(message string) string {
	if cmdData.AsEnvFile {
		return ""
	}

	return fmt.Sprintf("echo '%s'", message)
}

func skipLine(message string) string {
	return commentLine(fmt.Sprintf("skip %s", message))
}

func commentLine(message string) string {
	var commentSign = "#"
	if !cmdData.AsEnvFile && cmdData.Shell == "cmdexe" {
		commentSign = "::"
	}

	return fmt.Sprintf("%s %s", commentSign, message)
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

	var f *os.File
	if cmdData.OutputFilePath != "" {
		f, err = os.OpenFile(cmdData.OutputFilePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return "", err
		}
	} else {
		var tempFilePattern string
		if cmdData.AsFile {
			tempFilePattern = fmt.Sprintf("source_%d_*", time.Now().Unix())
			switch cmdData.Shell {
			case "cmdexe":
				tempFilePattern += ".bat"
			case "powershell":
				tempFilePattern += ".ps1"
			}
		} else {
			tempFilePattern = fmt.Sprintf(".env_%d_*", time.Now().Unix())
		}

		f, err = ioutil.TempFile(sourceDir, tempFilePattern)
		if err != nil {
			return "", err
		}
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return "", err
	}

	return f.Name(), nil
}
