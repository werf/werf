package ci_env

import (
	"bytes"
	"context"
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
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/determinism_inspector"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	AsFile         bool
	AsEnvFile      bool
	OutputFilePath string
	Shell          string
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
	common.SetupDisableDeterminism(&commonCmdData, cmd)
	common.SetupNonStrictDeterminismInspection(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupDockerConfig(&commonCmdData, cmd, "Command will copy specified or default (~/.docker) config to the temporary directory and may perform additional login with new config.")

	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.AsFile, "as-file", "", common.GetBoolEnvironmentDefaultFalse("WERF_AS_FILE"), "Create the script and print the path for sourcing (default $WERF_AS_FILE).")
	cmd.Flags().BoolVarP(&cmdData.AsEnvFile, "as-env-file", "", common.GetBoolEnvironmentDefaultFalse("WERF_AS_ENV_FILE"), "Create the .env file and print the path for sourcing (default $WERF_AS_ENV_FILE).")
	cmd.Flags().StringVarP(&cmdData.OutputFilePath, "output-file-path", "o", os.Getenv("WERF_OUTPUT_FILE_PATH"), "Write to custom file (default $WERF_OUTPUT_FILE_PATH).")
	cmd.Flags().StringVarP(&cmdData.Shell, "shell", "", os.Getenv("WERF_SHELL"), "Set to cmdexe, powershell or use the default behaviour that is compatible with any unix shell (default $WERF_SHELL).")

	return cmd
}

func runCIEnv(cmd *cobra.Command, args []string) error {
	logboek.SetAcceptedLevel(level.Error)

	ctx := common.BackgroundContext()

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := determinism_inspector.Init(determinism_inspector.InspectionOptions{NonStrict: *commonCmdData.NonStrictDeterminismInspection}); err != nil {
		return err
	}

	if err := git_repo.Init(); err != nil {
		return err
	}

	if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
		return err
	}

	dockerConfig, err := generateSessionDockerConfigDir(ctx)
	if err != nil {
		return err
	}

	if err := docker.Init(ctx, dockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return err
	}
	ctx = ctxWithDockerCli

	switch cmdData.Shell {
	case "", "default", "cmdexe", "powershell":
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided shell '%s' not supported", cmdData.Shell)
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

func generateGitlabEnvs(ctx context.Context, w io.Writer, dockerConfig string) error {
	ciRegistryImageEnv := os.Getenv("CI_REGISTRY_IMAGE")
	ciJobTokenEnv := os.Getenv("CI_JOB_TOKEN")

	var repo, repoImplementation string
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
			projectDir, err := common.GetProjectDir(&commonCmdData)
			if err != nil {
				return fmt.Errorf("getting project dir failed: %s", err)
			}

			localGitRepo, err := git_repo.OpenLocalRepo("own", projectDir)
			if err != nil {
				return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
			}

			werfConfig, err := common.GetOptionalWerfConfig(ctx, projectDir, &commonCmdData, localGitRepo, config.WerfConfigOptions{LogRenderedFilePath: true, DisableDeterminism: *commonCmdData.DisableDeterminism, Env: *commonCmdData.Environment})
			if err != nil {
				return fmt.Errorf("unable to load werf config: %s", err)
			}

			if werfConfig != nil {
				repo = fmt.Sprintf("%s/werf", ciRegistryImageEnv)
			}
		}

		if werfRepoEnv == "" || strings.HasPrefix(werfRepoEnv, ciRegistryEnv) {
			repoImplementation = docker_registry.GitLabRegistryImplementationName
		}
	}

	if doLogin {
		err := docker.Login(ctx, imagesUsername, imagesPassword, ciRegistryImageEnv)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %s", ciRegistryImageEnv, err)
		}
	}

	writeHeader(w, "DOCKER CONFIG", false)
	writeEnv(w, "DOCKER_CONFIG", dockerConfig, true)

	writeHeader(w, "REPO", true)
	writeEnv(w, "WERF_REPO", repo, false)
	if repoImplementation != "" {
		writeEnv(w, "WERF_REPO_IMPLEMENTATION", repoImplementation, false)
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

func generateGithubEnvs(ctx context.Context, w io.Writer, dockerConfig string) error {
	githubRegistry := "docker.pkg.github.com"
	ciGithubToken := os.Getenv("GITHUB_TOKEN")
	ciGithubActor := os.Getenv("GITHUB_ACTOR")
	if ciGithubActor != "" && ciGithubToken != "" {
		err := docker.Login(ctx, ciGithubActor, ciGithubToken, githubRegistry)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %s", githubRegistry, err)
		}
	}

	ciGithubOwnerWithProject := os.Getenv("GITHUB_REPOSITORY")
	ciGithubDockerPackage := strings.ToLower(ciGithubOwnerWithProject)

	var repo string
	if ciGithubDockerPackage != "" {
		projectDir, err := common.GetProjectDir(&commonCmdData)
		if err != nil {
			return fmt.Errorf("getting project dir failed: %s", err)
		}

		localGitRepo, err := git_repo.OpenLocalRepo("own", projectDir)
		if err != nil {
			return fmt.Errorf("unable to open local repo %s: %s", projectDir, err)
		}

		werfConfig, err := common.GetOptionalWerfConfig(ctx, projectDir, &commonCmdData, localGitRepo, config.WerfConfigOptions{LogRenderedFilePath: true, DisableDeterminism: *commonCmdData.DisableDeterminism, Env: *commonCmdData.Environment})
		if err != nil {
			return fmt.Errorf("unable to load werf config: %s", err)
		}

		if werfConfig != nil {
			projectRepo := fmt.Sprintf("%s/%s", githubRegistry, ciGithubDockerPackage)
			repo = fmt.Sprintf("%s/%s-werf", projectRepo, werfConfig.Meta.Project)
		}
	}

	writeHeader(w, "DOCKER CONFIG", false)
	writeEnv(w, "DOCKER_CONFIG", dockerConfig, true)

	writeHeader(w, "REPO", true)
	writeEnv(w, "WERF_REPO", repo, false)

	writeHeader(w, "DEPLOY", true)
	var projectGit string
	if ciGithubOwnerWithProject != "" {
		projectGit = fmt.Sprintf("project.werf.io/git=%s", fmt.Sprintf("https://github.com/%s", ciGithubOwnerWithProject))
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_PROJECT_GIT", projectGit, false)

	var ciCommit string
	ciCommitShaEnv := os.Getenv("GITHUB_SHA")
	if ciCommitShaEnv != "" {
		ciCommit = fmt.Sprintf("ci.werf.io/commit=%s", ciCommitShaEnv)
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_CI_COMMIT", ciCommit, false)

	var workflowRunUrl string
	ciWorkflowRunIdEnv := os.Getenv("GITHUB_RUN_ID")
	if ciGithubOwnerWithProject != "" && ciWorkflowRunIdEnv != "" {
		workflowRunUrl = fmt.Sprintf("github.ci.werf.io/workflow-run-url=%s", fmt.Sprintf("https://github.com/%s/actions/runs/%s", ciGithubOwnerWithProject, ciWorkflowRunIdEnv))
	}
	writeEnv(w, "WERF_ADD_ANNOTATION_GITHUB_ACTIONS_RUN_URL", workflowRunUrl, false)

	writeHeader(w, "CLEANUP", true)
	writeEnv(w, "WERF_REPO_GITHUB_TOKEN", ciGithubToken, false)

	if err := generateOther(w); err != nil {
		return err
	}

	return nil
}

func generateSessionDockerConfigDir(ctx context.Context) (string, error) {
	dockerConfigPath := *commonCmdData.DockerConfig
	if *commonCmdData.DockerConfig == "" {
		dockerConfigPath = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	dockerConfigDir, err := tmp_manager.CreateDockerConfigDir(ctx, dockerConfigPath)
	if err != nil {
		return "", fmt.Errorf("unable to create tmp docker config: %s", err)
	}

	return dockerConfigDir, nil
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
	var commentSign = "#"
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
