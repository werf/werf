package cmd_factory

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/werf"
)

type CmdData struct {
	PullUsername string
	PullPassword string

	IntrospectBeforeError bool
	IntrospectAfterError  bool
}

func NewCmdWithData(cmdData *CmdData, commonCmdData *common.CmdData) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [IMAGE_NAME...]",
		Short: "Build stages",
		Example: `  # Build stages of all images from werf.yaml, built stages will be placed locally
  $ werf stages build --stages-storage :local

  # Build stages of image 'backend' from werf.yaml
  $ werf stages build --stages-storage :local backend

  # Build and enable drop-in shell session in the failed assembly container in the case when an error occurred
  $ werf build --stages-storage :local --introspect-error

  # Set --stages-storage default value using $WERF_STAGES_STORAGE param
  $ export WERF_STAGES_STORAGE=:local
  $ werf build`,
		Long: common.GetLongCommandDescription(`Build stages for images described in the werf.yaml.

The result of build command are built stages pushed into the specified stages storage (or locally in the case when --stages-storage=:local).

If one or more IMAGE_NAME parameters specified, werf will build only these images stages from werf.yaml`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfDebugAnsibleArgs),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error {
				return runStagesBuild(cmdData, commonCmdData, args)
			})
		},
	}

	common.SetupDir(commonCmdData, cmd)
	common.SetupTmpDir(commonCmdData, cmd)
	common.SetupHomeDir(commonCmdData, cmd)
	common.SetupSSHKey(commonCmdData, cmd)

	common.SetupStagesStorage(commonCmdData, cmd)
	common.SetupStagesStorageLock(commonCmdData, cmd)
	common.SetupDockerConfig(commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified stages storage, to pull base images")
	common.SetupInsecureRegistry(commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(commonCmdData, cmd)

	common.SetupIntrospectStage(commonCmdData, cmd)

	common.SetupLogOptions(commonCmdData, cmd)
	common.SetupLogProjectDir(commonCmdData, cmd)

	cmd.Flags().BoolVarP(&cmdData.IntrospectAfterError, "introspect-error", "", false, "Introspect failed stage in the state, right after running failed assembly instruction")
	cmd.Flags().BoolVarP(&cmdData.IntrospectBeforeError, "introspect-before-error", "", false, "Introspect failed stage in the clean state, before running all assembly instructions of the stage")

	return cmd
}

func runStagesBuild(cmdData *CmdData, commonCmdData *common.CmdData, imagesToProcess []string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := shluz.Init(filepath.Join(werf.GetServiceDir(), "locks")); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{InsecureRegistry: *commonCmdData.InsecureRegistry, SkipTlsVerifyRegistry: *commonCmdData.SkipTlsVerifyRegistry}); err != nil {
		return err
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	common.ProcessLogProjectDir(commonCmdData, projectDir)

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, true)
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	for _, imageToProcess := range imagesToProcess {
		if !werfConfig.HasImage(imageToProcess) {
			return fmt.Errorf("specified image %s is not defined in werf.yaml", logging.ImageLogName(imageToProcess, false))
		}
	}

	projectTmpDir, err := tmp_manager.CreateProjectDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	stagesStorage, err := common.GetStagesStorage(commonCmdData)
	if err != nil {
		return err
	}

	_, err = common.GetStagesStorageLock(commonCmdData)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*commonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.LogWarnF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	introspectOptions, err := common.GetIntrospectOptions(commonCmdData, werfConfig)
	if err != nil {
		return err
	}

	opts := build.BuildStagesOptions{
		ImageBuildOptions: image.BuildOptions{
			IntrospectAfterError:  cmdData.IntrospectAfterError,
			IntrospectBeforeError: cmdData.IntrospectBeforeError,
		},
		IntrospectOptions: introspectOptions,
	}

	logboek.LogOptionalLn()
	c := build.NewConveyor(werfConfig, imagesToProcess, projectDir, projectTmpDir, ssh_agent.SSHAuthSock)
	defer c.Terminate()

	if err = c.BuildStages(stagesStorage, opts); err != nil {
		return err
	}

	return nil
}
