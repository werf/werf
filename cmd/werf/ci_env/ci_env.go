package ci_env

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	TaggingStrategyStub string
	AsFile              bool
	AsEnvFile           bool
	OutputFilePath      string
	Shell               string
	AllowRegistryLogin  bool
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
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
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupDockerConfig(&commonCmdData, cmd, "Command will copy specified or default (~/.docker) config to the temporary directory and may perform additional login with new config.")

	commonCmdData.SetupPlatform(cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.AllowRegistryLogin, "login-to-registry", "", util.GetBoolEnvironmentDefaultTrue("WERF_LOGIN_TO_REGISTRY"), "Log in to CI-specific registry automatically if possible (default $WERF_LOGIN_TO_REGISTRY).")
	cmd.Flags().BoolVarP(&cmdData.AsFile, "as-file", "", util.GetBoolEnvironmentDefaultFalse("WERF_AS_FILE"), "Create the script and print the path for sourcing (default $WERF_AS_FILE).")
	cmd.Flags().BoolVarP(&cmdData.AsEnvFile, "as-env-file", "", util.GetBoolEnvironmentDefaultFalse("WERF_AS_ENV_FILE"), "Create the .env file and print the path for sourcing (default $WERF_AS_ENV_FILE).")
	cmd.Flags().StringVarP(&cmdData.OutputFilePath, "output-file-path", "o", os.Getenv("WERF_OUTPUT_FILE_PATH"), "Write to custom file (default $WERF_OUTPUT_FILE_PATH).")
	cmd.Flags().StringVarP(&cmdData.Shell, "shell", "", os.Getenv("WERF_SHELL"), "Set to cmdexe, powershell or use the default behaviour that is compatible with any unix shell (default $WERF_SHELL).")
	cmd.Flags().StringVarP(&cmdData.TaggingStrategyStub, "tagging-strategy", "", "", `stub`)
	cmd.Flag("tagging-strategy").Hidden = true

	return cmd
}

func runCIEnv(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	logboek.SetAcceptedLevel(level.Error)

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
		return err
	}

	dockerConfig, err := generateSessionDockerConfigDir(ctx)
	if err != nil {
		return err
	}

	// FIXME(multiarch): do not initialize platform in backend here
	// FIXME(multiarch): why docker initialization here? what if buildah backend enabled?
	opts := docker.InitOptions{
		DockerConfigDir: dockerConfig,
		ClaimPlatforms:  commonCmdData.GetPlatform(),
		Verbose:         *commonCmdData.LogVerbose,
		Debug:           *commonCmdData.LogDebug,
	}
	if err := docker.Init(ctx, opts); err != nil {
		return fmt.Errorf("docker init failed in dir %q: %w", dockerConfig, err)
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return err
	}
	ctx = ctxWithDockerCli

	if cmdData.TaggingStrategyStub != "" && cmdData.AsFile {
		return errors.New("unknown flag: --tagging-strategy")
	}

	switch cmdData.Shell {
	case "", "default", "cmdexe", "powershell":
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided shell %q not supported", cmdData.Shell)
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
		err := generateGithubEnvs(ctx, w, dockerConfig)
		if err != nil {
			if !cmdData.AsFile && !cmdData.AsEnvFile {
				writeError(w, err.Error())
			}
			return err
		}
	case "gitlab":
		err := generateGitlabEnvs(ctx, w, dockerConfig)
		if err != nil {
			if !cmdData.AsFile && !cmdData.AsEnvFile {
				writeError(w, err.Error())
			}
			return err
		}
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided ci system %q not supported", ciSystem)
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

func generateGitlabEnvs(ctx context.Context, w io.Writer, dockerConfig string) error {
	ciRegistryImageEnv := os.Getenv("CI_REGISTRY_IMAGE")
	ciJobTokenEnv := os.Getenv("CI_JOB_TOKEN")

	var repo, repoContainerRegistry string
	var imagesUsername, imagesPassword string
	var doLogin bool
	if ciRegistryImageEnv != "" {
		if ciJobTokenEnv != "" {
			imagesUsername = "gitlab-ci-token"
			imagesPassword = ciJobTokenEnv
			doLogin = true
		}

		ciRegistryEnv := os.Getenv("CI_REGISTRY")
		werfRepoEnv := os.Getenv("WERF_REPO")

		if werfRepoEnv == "" {
			repo = ciRegistryImageEnv
		}

		if werfRepoEnv == "" || strings.HasPrefix(werfRepoEnv, ciRegistryEnv) {
			repoContainerRegistry = docker_registry.GitLabRegistryImplementationName
		}
	}

	if doLogin && cmdData.AllowRegistryLogin {
		err := docker.Login(ctx, imagesUsername, imagesPassword, ciRegistryImageEnv)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %w", ciRegistryImageEnv, err)
		}
	}

	writeHeader(w, "DOCKER CONFIG", false)
	writeEnv(w, "DOCKER_CONFIG", dockerConfig, true)

	writeHeader(w, "REPO", true)
	writeEnv(w, "WERF_REPO", repo, false)
	if repoContainerRegistry != "" {
		writeEnv(w, "WERF_REPO_CONTAINER_REGISTRY", repoContainerRegistry, false)
	}

	writeHeader(w, "DEPLOY", true)
	writeEnv(w, "WERF_ENV", os.Getenv("CI_ENVIRONMENT_SLUG"), false)

	var releaseChannel string
	trdlUseWerfGroupChannel := os.Getenv("TRDL_USE_WERF_GROUP_CHANNEL")
	if trdlUseWerfGroupChannel != "" {
		releaseChannel = fmt.Sprintf("werf.io/release-channel=%s", trdlUseWerfGroupChannel)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_WERF_RELEASE_CHANNEL", releaseChannel, true)

	var projectGit string
	ciProjectUrlEnv := os.Getenv("CI_PROJECT_URL")
	if ciProjectUrlEnv != "" {
		projectGit = fmt.Sprintf("project.werf.io/git=%s", ciProjectUrlEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_PROJECT_GIT", projectGit, true)

	var ciCommit string
	ciCommitShaEnv := os.Getenv("CI_COMMIT_SHA")
	if ciCommitShaEnv != "" {
		ciCommit = fmt.Sprintf("ci.werf.io/commit=%s", ciCommitShaEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_CI_COMMIT", ciCommit, true)

	var ciGitTag string
	ciCommitTag := os.Getenv("CI_COMMIT_TAG")
	if ciCommitTag != "" {
		ciGitTag = fmt.Sprintf("ci.werf.io/tag=%s", ciCommitTag)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_CI_GIT_TAG", ciGitTag, true)

	var gitlabCIPipelineUrl string
	ciPipelineIdEnv := os.Getenv("CI_PIPELINE_ID")
	if ciProjectUrlEnv != "" && ciPipelineIdEnv != "" {
		gitlabCIPipelineUrl = fmt.Sprintf("gitlab.ci.werf.io/pipeline-url=%s/pipelines/%s", ciProjectUrlEnv, ciPipelineIdEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL", gitlabCIPipelineUrl, true)

	var gitlabCiJobUrl string
	ciJobIdEnv := os.Getenv("CI_JOB_ID")
	if ciProjectUrlEnv != "" && os.Getenv("CI_JOB_ID") != "" {
		gitlabCiJobUrl = fmt.Sprintf("gitlab.ci.werf.io/job-url=%s/-/jobs/%s", ciProjectUrlEnv, ciJobIdEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL", gitlabCiJobUrl, true)

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
	writeEnv(w, "WERF_LOG_TERMINAL_WIDTH", "130", false)

	return nil
}

func generateGithubEnvs(ctx context.Context, w io.Writer, dockerConfig string) error {
	ciGithubOwnerWithProject := os.Getenv("GITHUB_REPOSITORY")
	ciGithubDockerPackage := strings.ToLower(ciGithubOwnerWithProject)
	defaultRegistry := docker_registry.GitHubPackagesRegistryAddress
	defaultRepo, err := generateGithubDefaultRepo(ctx, defaultRegistry, ciGithubDockerPackage)
	if err != nil {
		return fmt.Errorf("unable to generate default repo: %w", err)
	}

	// TODO: legacy, delete when upgrading to v1.3
	registryToLogin := defaultRegistry
	customRepo := os.Getenv("WERF_REPO")
	gitHubPackagesRegistryAddressOld := "docker.pkg.github.com"
	if strings.HasPrefix(customRepo, gitHubPackagesRegistryAddressOld) {
		registryToLogin = gitHubPackagesRegistryAddressOld
	}

	ciGithubActor := os.Getenv("GITHUB_ACTOR")
	ciGithubToken := os.Getenv("GITHUB_TOKEN")
	if ciGithubActor != "" && ciGithubToken != "" && cmdData.AllowRegistryLogin {
		err := docker.Login(ctx, ciGithubActor, ciGithubToken, registryToLogin)
		if err != nil {
			return fmt.Errorf("unable to login into docker registry %s: %w", defaultRegistry, err)
		}
	}

	writeHeader(w, "DOCKER CONFIG", false)
	writeEnv(w, "DOCKER_CONFIG", dockerConfig, true)

	writeHeader(w, "REPO", true)
	writeEnv(w, "WERF_REPO", defaultRepo, false)

	writeHeader(w, "DEPLOY", true)

	var releaseChannel string
	trdlUseWerfGroupChannel := os.Getenv("TRDL_USE_WERF_GROUP_CHANNEL")
	if trdlUseWerfGroupChannel != "" {
		releaseChannel = fmt.Sprintf("werf.io/release-channel=%s", trdlUseWerfGroupChannel)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_WERF_RELEASE_CHANNEL", releaseChannel, true)

	var projectGit string
	if ciGithubOwnerWithProject != "" {
		projectGit = fmt.Sprintf("project.werf.io/git=%s", fmt.Sprintf("https://github.com/%s", ciGithubOwnerWithProject))
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_PROJECT_GIT", projectGit, true)

	var ciCommit string
	ciCommitShaEnv := os.Getenv("GITHUB_SHA")
	if ciCommitShaEnv != "" {
		ciCommit = fmt.Sprintf("ci.werf.io/commit=%s", ciCommitShaEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_CI_COMMIT", ciCommit, true)

	var ciGitTag string
	ciRefType := os.Getenv("GITHUB_REF_TYPE")
	ciRefName := os.Getenv("GITHUB_REF_NAME")
	if ciRefType == "tag" && ciRefName != "" {
		ciGitTag = fmt.Sprintf("ci.werf.io/tag=%s", ciRefName)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_CI_GIT_TAG", ciGitTag, true)

	var workflowRunUrl string
	ciWorkflowRunIdEnv := os.Getenv("GITHUB_RUN_ID")
	if ciGithubOwnerWithProject != "" && ciWorkflowRunIdEnv != "" {
		workflowRunUrl = fmt.Sprintf("github.ci.werf.io/workflow-run-url=%s", fmt.Sprintf("https://github.com/%s/actions/runs/%s", ciGithubOwnerWithProject, ciWorkflowRunIdEnv))
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_GITHUB_ACTIONS_RUN_URL", workflowRunUrl, true)

	writeHeader(w, "CLEANUP", true)
	writeEnv(w, "WERF_REPO_GITHUB_TOKEN", ciGithubToken, false)

	if err := generateOther(w); err != nil {
		return err
	}

	return nil
}

func generateGithubDefaultRepo(ctx context.Context, defaultRegistry, ciGithubDockerPackage string) (string, error) {
	var defaultRepo string
	if ciGithubDockerPackage != "" {
		giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
		if err != nil {
			return "", err
		}

		_, werfConfig, err := common.GetOptionalWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, true))
		if err != nil {
			return "", fmt.Errorf("unable to load werf config: %w", err)
		}

		if werfConfig != nil {
			projectRepo := fmt.Sprintf("%s/%s", defaultRegistry, ciGithubDockerPackage)
			defaultRepo = fmt.Sprintf("%s/%s", projectRepo, werfConfig.Meta.Project)
		}
	}

	return defaultRepo, nil
}

func generateSessionDockerConfigDir(ctx context.Context) (string, error) {
	dockerConfigPath := *commonCmdData.DockerConfig
	if *commonCmdData.DockerConfig == "" {
		dockerConfigPath = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	dockerConfigDir, err := tmp_manager.CreateDockerConfigDir(ctx, dockerConfigPath)
	if err != nil {
		return "", fmt.Errorf("unable to create tmp docker config: %w", err)
	}

	return dockerConfigDir, nil
}

func generateOther(w io.Writer) error {
	writeHeader(w, "OTHER", true)
	writeEnv(w, "WERF_LOG_COLOR_MODE", "on", false)
	writeEnv(w, "WERF_LOG_PROJECT_DIR", "1", false)
	writeEnv(w, "WERF_ENABLE_PROCESS_EXTERMINATOR", "1", false)
	writeEnv(w, "WERF_LOG_TERMINAL_WIDTH", "130", false)

	return nil
}

func writeError(w io.Writer, errMsg string) {
	if *commonCmdData.LogVerbose {
		writeEcho(w, "")
		writeEcho(w, "Error: "+errMsg)
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
			writeEcho(w, "")
		}

		writeEcho(w, headerLine)
	}
}

func writeEnv(w io.Writer, key, value string, override bool) {
	envLine := envLine(key, value)

	if !override && os.Getenv(key) != "" {
		skipLine := skipLine(fmt.Sprintf("%s (%s)", envLine, os.Getenv(key)))
		writeLn(w, skipLine)

		if *commonCmdData.LogVerbose {
			writeEcho(w, skipLine)
		}

		return
	}

	if value == "" {
		envLine = commentLine(envLine)
	}

	writeLn(w, envLine)

	if *commonCmdData.LogVerbose {
		writeEcho(w, envLine)
	}
}

func writeEcho(w io.Writer, message string) {
	if cmdData.AsEnvFile {
		return
	}

	writeLn(w, fmt.Sprintf("echo '%s'", message))
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

func skipLine(message string) string {
	return commentLine(fmt.Sprintf("skip %s", message))
}

func commentLine(message string) string {
	commentSign := "#"
	if !cmdData.AsEnvFile && cmdData.Shell == "cmdexe" {
		commentSign = "::"
	}

	return fmt.Sprintf("%s %s", commentSign, message)
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
		f, err = os.OpenFile(cmdData.OutputFilePath, os.O_RDWR|os.O_CREATE, 0o755)
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
