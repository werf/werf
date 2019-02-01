package generate_chart

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/docker_authorizer"
	"github.com/flant/werf/pkg/deploy"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/ssh_agent"
	"github.com/flant/werf/pkg/true_git"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CmdData struct {
	RegistryUsername string
	RegistryPassword string

	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-chart PATH",
		Short: "Generate Werf chart which will contain a valid Helm chart to the specified path",
		Long: common.GetLongCommandDescription(`Generate Werf chart which will contain a valid Helm chart to the specified path.

Werf will generate additional values files, templates Chart.yaml and other files specific to the Werf chart. The result is a valid Helm chart.`),
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey, common.WerfDockerConfig, common.WerfHome, common.WerfTmp),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runGenerateChart(args[0]); err != nil {
				return fmt.Errorf("generate-chart failed: %s", err)
			}

			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.RegistryUsername, "registry-username", "", "", "Docker registry username")
	cmd.Flags().StringVarP(&CmdData.RegistryPassword, "registry-password", "", "", "Docker registry password")

	cmd.Flags().StringArrayVarP(&CmdData.Values, "values", "", []string{}, "Additional helm values")
	cmd.Flags().StringArrayVarP(&CmdData.SecretValues, "secret-values", "", []string{}, "Additional helm secret values")
	cmd.Flags().StringArrayVarP(&CmdData.Set, "set", "", []string{}, "Additional helm sets")
	cmd.Flags().StringArrayVarP(&CmdData.SetString, "set-string", "", []string{}, "Additional helm STRING sets")

	common.SetupTag(&CommonCmdData, cmd)
	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupNamespace(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)

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

	if err := true_git.Init(); err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh-agent: %s", err)
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

	imagesRepo := common.GetOptionalImagesRepo(werfConfig.Meta.Project, &CommonCmdData)
	withoutRegistry := true

	if imagesRepo != "" {
		withoutRegistry = false

		var err error

		projectTmpDir, err := project_tmp_dir.Get()
		if err != nil {
			return fmt.Errorf("getting project tmp dir failed: %s", err)
		}
		defer project_tmp_dir.Release(projectTmpDir)

		dockerAuthorizer, err := docker_authorizer.GetCommonDockerAuthorizer(projectTmpDir, CmdData.RegistryUsername, CmdData.RegistryPassword)
		if err != nil {
			return err
		}

		if err := dockerAuthorizer.Login(imagesRepo); err != nil {
			return fmt.Errorf("docker login failed: %s", err)
		}
	}

	if imagesRepo == "" {
		imagesRepo = "IMAGES_REPO"
	}

	environment := *CommonCmdData.Environment
	if environment == "" {
		environment = "ENV"
	}

	namespace, err := common.GetKubernetesNamespace(*CommonCmdData.Namespace, environment, werfConfig)
	if err != nil {
		return err
	}

	tag, err := common.GetDeployTag(&CommonCmdData)
	if err != nil {
		return err
	}

	localGit := &git_repo.Local{Path: projectDir, GitDir: filepath.Join(projectDir, ".git")}

	images := deploy.GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, withoutRegistry)

	serviceValues, err := deploy.GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, localGit, images, deploy.ServiceValuesOptions{})
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

	werfChart, err := deploy.PrepareWerfChart(targetPath, werfConfig.Meta.Project, projectDir, m, CmdData.Values, CmdData.SecretValues, CmdData.Set, CmdData.SetString, serviceValues)
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
