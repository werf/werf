package generate_chart

import (
	"fmt"
	"os"

	"github.com/flant/logboek"
	helm_common "github.com/flant/werf/cmd/werf/helm/common"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CmdData struct {
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-chart PATH",
		Short: "Generate Werf chart which will contain a valid Helm chart to the specified path",
		Long: common.GetLongCommandDescription(`Generate Werf chart which will contain a valid Helm chart to the specified path.

Werf will generate additional values files, templates Chart.yaml and other files specific to the Werf chart. The result is a valid Helm chart`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ValidateArgumentCount(1, args, cmd); err != nil {
				return err
			}

			return runGenerateChart(args[0])
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	common.SetupTag(&CommonCmdData, cmd)
	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupNamespace(&CommonCmdData, cmd)

	common.SetupSecretValues(&CommonCmdData, cmd)
	common.SetupIgnoreSecretKey(&CommonCmdData, cmd)

	common.SetupStagesStorage(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified stages storage and images repo")
	common.SetupInsecureRepo(&CommonCmdData, cmd)

	return cmd
}

func runGenerateChart(targetPath string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := deploy.Init(deploy.InitOptions{WithoutHelm: true}); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logboek.GetOutStream(), Err: logboek.GetErrStream()}); err != nil {
		return err
	}

	if err := docker_registry.Init(docker_registry.Options{AllowInsecureRepo: *CommonCmdData.InsecureRepo}); err != nil {
		return err
	}

	if err := docker.Init(*CommonCmdData.DockerConfig); err != nil {
		return err
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("bad config: %s", err)
	}

	imagesRepo := common.GetOptionalImagesRepo(werfConfig.Meta.Project, &CommonCmdData)
	withoutRepo := true
	if imagesRepo != "" {
		withoutRepo = false
	}

	imagesRepo = helm_common.GetImagesRepoOrStub(imagesRepo)

	environment := helm_common.GetEnvironmentOrStub(*CommonCmdData.Environment)

	namespace, err := common.GetKubernetesNamespace(*CommonCmdData.Namespace, environment, werfConfig)
	if err != nil {
		return err
	}

	tag, tagStrategy, err := helm_common.GetTagOrStub(&CommonCmdData)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.LogErrorF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	images := deploy.GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, withoutRepo)

	serviceValues, err := deploy.GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagStrategy, images, deploy.ServiceValuesOptions{Env: environment})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	m, err := deploy.GetSafeSecretManager(projectDir, *CommonCmdData.SecretValues, *CommonCmdData.IgnoreSecretKey)
	if err != nil {
		return err
	}

	targetPath = util.ExpandPath(targetPath)

	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		logboek.LogF("Removing existing %s\n", targetPath)
		err = os.RemoveAll(targetPath)
		if err != nil {
			return err
		}
	}

	werfChart, err := deploy.PrepareWerfChart(targetPath, werfConfig.Meta.Project, projectDir, environment, m, *CommonCmdData.SecretValues, serviceValues)
	if err != nil {
		return err
	}

	err = werfChart.Save()
	if err != nil {
		return fmt.Errorf("unable to save werf chart: %s", err)
	}

	logboek.LogF("Generated werf chart: %s\n", targetPath)

	return nil
}
