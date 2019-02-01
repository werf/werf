package deploy

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/flant/kubedog/pkg/kube"
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
	"github.com/flant/werf/pkg/werf"
	"github.com/spf13/cobra"
)

var CmdData struct {
	Timeout int

	Values       []string
	SecretValues []string
	Set          []string
	SetString    []string

	RegistryUsername string
	RegistryPassword string
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy application into Kubernetes",
		Long: common.GetLongCommandDescription(`Deploy application into Kubernetes.

Command will create Helm Release and wait until all resources of the release are become ready.

Deploy needs the same parameters as push to construct image names: repo and tags. Docker images names are constructed from paramters as IMAGES_REPO/IMAGE_NAME:TAG. Deploy will fetch built image ids from Docker registry. So images should be published prior running deploy.

Helm chart directory .helm should exists and contain valid Helm chart.

Environment is a required param for the deploy by default, because it is needed to construct Helm Release name and Kubernetes Namespace. Either --env or WERF_DEPLOY_ENVIRONMENT should be specified for command.

Read more info about Helm chart structure, Helm Release name, Kubernetes Namespace and how to change it: https://flant.github.io/werf/reference/deploy/deploy_to_kubernetes.html`),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfSecretKey, common.WerfDockerConfig, common.WerfHome, common.WerfTmp),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			return common.LogRunningTime(func() error {
				err := runDeploy()
				if err != nil {
					return fmt.Errorf("deploy failed: %s", err)
				}

				return nil
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupSSHKey(&CommonCmdData, cmd)

	cmd.Flags().IntVarP(&CmdData.Timeout, "timeout", "t", 0, "Resources tracking timeout in seconds")

	cmd.Flags().StringArrayVarP(&CmdData.Values, "values", "", []string{}, "Additional helm values")
	cmd.Flags().StringArrayVarP(&CmdData.SecretValues, "secret-values", "", []string{}, "Additional helm secret values")
	cmd.Flags().StringArrayVarP(&CmdData.Set, "set", "", []string{}, "Additional helm sets")
	cmd.Flags().StringArrayVarP(&CmdData.SetString, "set-string", "", []string{}, "Additional helm STRING sets")

	cmd.Flags().StringVarP(&CmdData.RegistryUsername, "registry-username", "", "", "Docker registry username")
	cmd.Flags().StringVarP(&CmdData.RegistryPassword, "registry-password", "", "", "Docker registry password")

	common.SetupTag(&CommonCmdData, cmd)
	common.SetupEnvironment(&CommonCmdData, cmd)
	common.SetupRelease(&CommonCmdData, cmd)
	common.SetupNamespace(&CommonCmdData, cmd)
	common.SetupKubeContext(&CommonCmdData, cmd)

	common.SetupStagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)

	return cmd
}

func runDeploy() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	if err := deploy.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	if err := ssh_agent.Init(*CommonCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh-agent: %s", err)
	}

	kubeContext := common.GetKubeContext(*CommonCmdData.KubeContext)
	if err := kube.Init(kube.InitOptions{KubeContext: kubeContext}); err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}
	common.LogProjectDir(projectDir)

	projectTmpDir, err := project_tmp_dir.Get()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer project_tmp_dir.Release(projectTmpDir)

	werfConfig, err := common.GetWerfConfig(projectDir)
	if err != nil {
		return fmt.Errorf("cannot parse werf config: %s", err)
	}

	_, err = common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	imagesRepo, err := common.GetImagesRepo(werfConfig.Meta.Project, &CommonCmdData)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetCommonDockerAuthorizer(projectTmpDir, CmdData.RegistryUsername, CmdData.RegistryPassword)
	if err != nil {
		return err
	}

	if err := dockerAuthorizer.Login(imagesRepo); err != nil {
		return fmt.Errorf("docker login failed: %s", err)
	}

	tag, err := common.GetDeployTag(&CommonCmdData)
	if err != nil {
		return err
	}

	release, err := common.GetHelmRelease(*CommonCmdData.Release, *CommonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	namespace, err := common.GetKubernetesNamespace(*CommonCmdData.Namespace, *CommonCmdData.Environment, werfConfig)
	if err != nil {
		return err
	}

	logger.LogInfoF("Using Helm release name: %s\n", release)
	logger.LogInfoF("Using Kubernetes namespace: %s\n", namespace)

	localGit := &git_repo.Local{Path: projectDir, GitDir: filepath.Join(projectDir, ".git")}

	images := deploy.GetImagesInfoGetters(werfConfig.Images, imagesRepo, tag, false)

	serviceValues, err := deploy.GetServiceValues(werfConfig.Meta.Project, imagesRepo, namespace, tag, localGit, images, deploy.ServiceValuesOptions{})
	if err != nil {
		return fmt.Errorf("error creating service values: %s", err)
	}

	m, err := deploy.GetSafeSecretManager(projectDir, CmdData.SecretValues)
	if err != nil {
		return fmt.Errorf("cannot get project secret: %s", err)
	}

	werfChart, err := deploy.PrepareWerfChart(deploy.GetTmpWerfChartPath(werfConfig.Meta.Project), werfConfig.Meta.Project, projectDir, m, CmdData.Values, CmdData.SecretValues, CmdData.Set, CmdData.SetString, serviceValues)
	if err != nil {
		return err
	}

	return werfChart.Deploy(release, namespace, deploy.HelmChartOptions{
		CommonHelmOptions: deploy.CommonHelmOptions{KubeContext: kubeContext},
		Timeout:           time.Duration(CmdData.Timeout) * time.Second,
	})
}
