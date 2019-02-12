package generate_chart

import (
	"fmt"
	"os"

	helm_common "github.com/flant/werf/cmd/werf/helm/common"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CmdData struct {
	SecretValues []string
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-chart PATH",
		Short: "Generate Werf chart which will contain a valid Helm chart to the specified path",
		Long: common.GetLongCommandDescription(`Generate Werf chart which will contain a valid Helm chart to the specified path.

Werf will generate additional values files, templates Chart.yaml and other files specific to the Werf chart. The result is a valid Helm chart`),
		DisableFlagsInUseLine: true,
		Args: cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

	common.SetupStagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)
	common.SetupDockerConfig(&CommonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified stages storage and images repo")

	cmd.Flags().StringArrayVarP(&CmdData.SecretValues, "secret-values", "", []string{}, "Additional helm secret values")

	return cmd
}

func runGenerateChart(targetPath string) error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := deploy.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{Out: logger.GetOutStream(), Err: logger.GetErrStream()}); err != nil {
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
		return fmt.Errorf("cannot parse werf config: %s", err)
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

	tag, tagStrategy, err := common.GetDeployTag(&CommonCmdData)
	if err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logger.LogErrorF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	images := deploy.GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, withoutRepo)

	serviceValues, err := deploy.GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, tagStrategy, images)
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	m, err := deploy.GetSafeSecretManager(projectDir, CmdData.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	targetPath = util.ExpandPath(targetPath)

	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		logger.LogInfoF("Removing existing %s\n", targetPath)
		err = os.RemoveAll(targetPath)
		if err != nil {
			return err
		}
	}

	werfChart, err := deploy.PrepareWerfChart(targetPath, werfConfig.Meta.Project, projectDir, m, CmdData.SecretValues, serviceValues)
	if err != nil {
		return err
	}

	err = werfChart.Save()
	if err != nil {
		return fmt.Errorf("unable to save Werf chart: %s", err)
	}

	logger.LogServiceF("Generated Werf chart %s\n", targetPath)

	return nil
}
