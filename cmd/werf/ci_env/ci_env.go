package ci_env

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/shluz"

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

var CmdData struct {
	TaggingStrategy string
	Verbose         bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "ci-env CI_SYSTEM",
		DisableFlagsInUseLine: true,
		Short:                 "Generate werf environment variables for specified CI system",
		Long: `Generate werf environment variables for specified CI system.

Currently supported only GitLab CI`,
		Example: `  # Load generated werf environment variables on GitLab job runner
  $ source <(werf ci-env gitlab --tagging-strategy tag-or-branch)`,
		RunE: runCIEnv,
	}

	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command will copy specified or default (~/.docker) config to the temporary directory and may perform additional login with new config")
	common.SetupInsecureRegistry(&CommonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.TaggingStrategy, "tagging-strategy", "", "stages-signature", "stages-signature: always use '--tag-by-stages-signature' option to tag all published images by corresponding stages-signature; tag-or-branch: generate auto '--tag-git-branch' or '--tag-git-tag' tag by specified CI_SYSTEM environment variables")
	cmd.Flags().BoolVarP(&CmdData.Verbose, "verbose", "", false, "Generate echo command for each resulted script line")

	return cmd
}

func runCIEnv(cmd *cobra.Command, args []string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
		return err
	}

	switch CmdData.TaggingStrategy {
	case "tag-or-branch", "stages-signature":
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided tagging-strategy '%s' not supported", CmdData.TaggingStrategy)
	}

	ciSystem := args[0]

	switch ciSystem {
	case "gitlab":
		err := generateGitlabEnvs(CmdData.TaggingStrategy)
		if err != nil {
			fmt.Println()
			printError(err.Error())
		}
		return err
	default:
		common.PrintHelp(cmd)
		return fmt.Errorf("provided ci system '%s' not supported", ciSystem)
	}
}

func generateGitlabEnvs(taggingStrategy string) error {
	dockerConfigPath := *CommonCmdData.DockerConfig
	if *CommonCmdData.DockerConfig == "" {
		dockerConfigPath = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	tmp_manager.AutoGCEnabled = false

	dockerConfig, err := tmp_manager.CreateDockerConfigDir(dockerConfigPath)
	if err != nil {
		return fmt.Errorf("unable to create tmp docker config: %s", err)
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *CommonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *CommonCmdData.SkipTlsVerifyRegistry}); err != nil {
		return err
	}

	// Init with new docker config dir
	if err := docker.Init(dockerConfig); err != nil {
		return err
	}

	ciRegistryImage := os.Getenv("CI_REGISTRY_IMAGE")
	ciJobToken := os.Getenv("CI_JOB_TOKEN")

	var imagesUsername, imagesPassword string
	var doLogin bool
	if ciRegistryImage != "" && ciJobToken != "" {
		imagesUsername = "gitlab-ci-token"
		imagesPassword = ciJobToken
		doLogin = true
	}

	if doLogin {
		err := docker.Login(imagesUsername, imagesPassword, ciRegistryImage)
		if err != nil {
			return fmt.Errorf("unable to login into docker repo %s: %s", ciRegistryImage, err)
		}
	}

	printHeader("DOCKER CONFIG", false)
	printExportCommand("DOCKER_CONFIG", dockerConfig, true)

	printHeader("IMAGES REPO", true)
	printExportCommand("WERF_IMAGES_REPO", ciRegistryImage, false)

	printHeader("TAGGING", true)
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
			printExportCommand("WERF_TAG_GIT_TAG", slug.DockerTag(ciGitTag), false)
		}
		if ciGitBranch != "" {
			printExportCommand("WERF_TAG_GIT_BRANCH", slug.DockerTag(ciGitBranch), false)
		}

		if ciGitTag == "" && ciGitBranch == "" {
			return fmt.Errorf("none of enviroment variables $WERF_TAG_GIT_TAG=$CI_COMMIT_TAG or $WERF_TAG_GIT_BRANCH=$CI_COMMIT_REF_NAME for '%s' strategy are detected", CmdData.TaggingStrategy)
		}
	case "stages-signature":
		printExportCommand("WERF_TAG_BY_STAGES_SIGNATURE", "true", false)
	}

	printHeader("DEPLOY", true)
	printExportCommand("WERF_ENV", os.Getenv("CI_ENVIRONMENT_SLUG"), false)

	var projectGit string
	ciProjectUrlEnv := os.Getenv("CI_PROJECT_URL")
	if ciProjectUrlEnv != "" {
		projectGit = fmt.Sprintf("project.werf.io/git=%s", ciProjectUrlEnv)
	}
	printExportCommand("WERF_ADD_ANNOTATION_PROJECT_GIT", projectGit, false)

	var ciCommit string
	ciCommitShaEnv := os.Getenv("CI_COMMIT_SHA")
	if ciCommitShaEnv != "" {
		ciCommit = fmt.Sprintf("ci.werf.io/commit=%s", ciCommitShaEnv)
	}
	printExportCommand("WERF_ADD_ANNOTATION_CI_COMMIT", ciCommit, false)

	var gitlabCIPipelineUrl string
	ciPipelineIdEnv := os.Getenv("CI_PIPELINE_ID")
	if ciProjectUrlEnv != "" && ciPipelineIdEnv != "" {
		gitlabCIPipelineUrl = fmt.Sprintf("gitlab.ci.werf.io/pipeline-url=%s/pipelines/%s", ciProjectUrlEnv, ciPipelineIdEnv)
	}
	printExportCommand("WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL", gitlabCIPipelineUrl, false)

	var gitlabCiJobUrl string
	ciJobIdEnv := os.Getenv("CI_JOB_ID")
	if ciProjectUrlEnv != "" && os.Getenv("CI_JOB_ID") != "" {
		gitlabCiJobUrl = fmt.Sprintf("gitlab.ci.werf.io/job-url=%s/-/jobs/%s", ciProjectUrlEnv, ciJobIdEnv)
	}
	printExportCommand("WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL", gitlabCiJobUrl, false)

	cleanupConfig, err := getCleanupConfig()
	if err != nil {
		return fmt.Errorf("unable to get cleanup config: %s", err)
	}

	printHeader("IMAGE CLEANUP POLICIES", true)
	printExportCommand("WERF_GIT_TAG_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyLimit), false)
	printExportCommand("WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitTagStrategyExpiryDays), false)
	printExportCommand("WERF_GIT_COMMIT_STRATEGY_LIMIT", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyLimit), false)
	printExportCommand("WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS", fmt.Sprintf("%d", cleanupConfig.GitCommitStrategyExpiryDays), false)

	printHeader("OTHER", true)

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

	printExportCommand("WERF_LOG_COLOR_MODE", werfLogColorMode, false)
	printExportCommand("WERF_LOG_PROJECT_DIR", "1", false)
	printExportCommand("WERF_ENABLE_PROCESS_EXTERMINATOR", "1", false)
	printExportCommand("WERF_LOG_TERMINAL_WIDTH", "95", false)

	return nil
}

func printError(errMsg string) {
	if CmdData.Verbose {
		fmt.Println("echo")
		fmt.Printf("echo 'Error: %s'\n", errMsg)
	}

	fmt.Printf("exit 1\n")
	fmt.Println()
}

func printHeader(header string, withNewLine bool) {
	header = fmt.Sprintf("### %s", header)

	if withNewLine {
		fmt.Println()
	}
	fmt.Println(header)

	if CmdData.Verbose {
		if withNewLine {
			fmt.Println("echo")
		}
		echoHeader := fmt.Sprintf("echo '%s'", header)
		fmt.Println(echoHeader)
	}
}

func printExportCommand(key, value string, override bool) {
	if !override && os.Getenv(key) != "" {
		skipComment := fmt.Sprintf("# skip %s=\"%s\"", key, os.Getenv(key))
		fmt.Println(skipComment)

		if CmdData.Verbose {
			echoSkip := fmt.Sprintf("echo '%s'", skipComment)
			fmt.Println(echoSkip)
		}

		return
	}

	exportCommand := fmt.Sprintf("export %s=\"%s\"", key, value)
	if value == "" {
		exportCommand = fmt.Sprintf("# %s", exportCommand)
	}

	fmt.Println(exportCommand)

	if CmdData.Verbose {
		echoExportCommand := fmt.Sprintf("echo '%s'", exportCommand)
		fmt.Println(echoExportCommand)
	}
}

type CleanupConfig struct {
	GitTagStrategyLimit         int `yaml:"gitTagStrategyLimit"`
	GitTagStrategyExpiryDays    int `yaml:"gitTagStrategyExpiryDays"`
	GitCommitStrategyLimit      int `yaml:"gitCommitStrategyLimit"`
	GitCommitStrategyExpiryDays int `yaml:"gitCommitStrategyExpiryDays"`
}

func getCleanupConfig() (CleanupConfig, error) {
	configPath := filepath.Join(werf.GetHomeDir(), "config", "cleanup.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return CleanupConfig{
			GitTagStrategyLimit:         10,
			GitTagStrategyExpiryDays:    30,
			GitCommitStrategyLimit:      50,
			GitCommitStrategyExpiryDays: 30,
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
