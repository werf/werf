package bpd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/docker_authorizer"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	Repo       string
	WithStages bool

	PullUsername     string
	PullPassword     string
	PushUsername     string
	PushPassword     string
	RegistryUsername string
	RegistryPassword string

	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bpd [IMAGE_NAME...]",
		Short: "Build, push built images into Docker registry, then run deploy",
		Long: common.GetLongCommandDescription(`Build, push built images into Docker registry, then run deploy.

This is combined command, which allows more optimized pipelines in your CI, compared to running build, push and deploy commands separately. Common phases for these commands will be calculated only once.

Images will be tagged automatically with the names REPO/IMAGE_NAME:TAG. These tags will be deleted after push. See more info about images naming: https://flant.github.io/werf/reference/registry/image_naming.html.

The result of bpd command is complete application deployed into kubernetes with all needed images built in one command.

If one or more IMAGE_NAME parameters specified, werf will build and push only these images from werf.yaml.`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfAnsibleArgs, common.WerfDockerConfig, common.WerfIgnoreCIDockerAutologin, common.WerfHome, common.WerfTmp),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			if CmdData.PullUsername == "" {
				CmdData.PullUsername = CmdData.RegistryUsername
			}
			if CmdData.PullPassword == "" {
				CmdData.PullPassword = CmdData.RegistryPassword
			}
			if CmdData.PushUsername == "" {
				CmdData.PushUsername = CmdData.RegistryUsername
			}
			if CmdData.PushPassword == "" {
				CmdData.PushPassword = CmdData.RegistryPassword
			}

			return common.LogRunningTime(func() error {
				err := runBP(args)
				if err != nil {
					return fmt.Errorf("bpd failed: %s", err)
				}

				return nil
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.Repo, "repo", "", "", "Docker repository name to push images to. CI_REGISTRY_IMAGE will be used by default if available.")
	cmd.Flags().StringVarP(&CmdData.PullUsername, "pull-username", "", "", "Docker registry username to authorize pull of base images")
	cmd.Flags().StringVarP(&CmdData.PullPassword, "pull-password", "", "", "Docker registry password to authorize pull of base images")
	cmd.Flags().StringVarP(&CmdData.PushUsername, "push-username", "", "", "Docker registry username to authorize push to the docker repo")
	cmd.Flags().StringVarP(&CmdData.PushPassword, "push-password", "", "", "Docker registry password to authorize push to the docker repo")
	cmd.Flags().StringVarP(&CmdData.RegistryUsername, "registry-username", "", "", "Docker registry username to authorize pull of base images and push to the docker repo")
	cmd.Flags().StringVarP(&CmdData.RegistryUsername, "registry-password", "", "", "Docker registry password to authorize pull of base images and push to the docker repo")

	cmd.Flags().StringVarP(&CmdData.Repo, "repo", "", "", "Docker repository name to get images ids from. CI_REGISTRY_IMAGE will be used by default if available.")

	cmd.Flags().BoolVarP(&CmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
	cmd.Flags().BoolVarP(&CmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")

	common.SetupTag(&CommonCmdData, cmd)

	cmd.Flags().StringArrayVarP(&CmdData.Values, "values", "", []string{}, "Additional helm values")
	cmd.Flags().StringArrayVarP(&CmdData.SecretValues, "secret-values", "", []string{}, "Additional helm secret values")
	cmd.Flags().StringArrayVarP(&CmdData.Set, "set", "", []string{}, "Additional helm sets")
	cmd.Flags().StringArrayVarP(&CmdData.SetString, "set-string", "", []string{}, "Additional helm STRING sets")

	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupRelease(&CommonCmdData, cmd)
	common.SetupNamespace(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)

	return cmd
}

func runBP(imagesToProcess []string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	projectName := werfConfig.Meta.Project

	projectBuildDir, err := common.GetProjectBuildDir(projectName)
	if err != nil {
		return fmt.Errorf("getting project build dir failed: %s", err)
	}

	projectTmpDir, err := project_tmp_dir.Get()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer project_tmp_dir.Release(projectTmpDir)

	repo, err := common.GetRequiredRepoName(projectName, CmdData.Repo)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetBPDockerAuthorizer(projectTmpDir, CmdData.PullUsername, CmdData.PullPassword, CmdData.PushUsername, CmdData.PushPassword, repo)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logger.LogWarningF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	tagOpts, err := common.GetTagOptions(&CommonCmdData, projectDir)
	if err != nil {
		return err
	}

	buildOpts := build.BuildOptions{
		ImageBuildOptions: image.BuildOptions{
			IntrospectAfterError:  CmdData.IntrospectAfterError,
			IntrospectBeforeError: CmdData.IntrospectBeforeError,
		},
	}

	pushOpts := build.PushOptions{TagOptions: tagOpts, WithStages: CmdData.WithStages}

	c := build.NewConveyor(werfConfig, imagesToProcess, projectDir, projectBuildDir, projectTmpDir, ssh_agent.SSHAuthSock, dockerAuthorizer)
	if err = c.BP(repo, buildOpts, pushOpts); err != nil {
		return err
	}

	return nil
}
