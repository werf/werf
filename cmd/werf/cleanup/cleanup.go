package cleanup

import (
	"fmt"
	"path"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/docker_authorizer"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"

	"github.com/spf13/cobra"
)

var CmdData struct {
	RegistryUsername string
	RegistryPassword string

	WithoutKube bool

	DryRun bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cleanup",
		DisableFlagsInUseLine: true,
		Short:                 "Cleanup project images registry and stages storage",
		Long:                  common.GetLongCommandDescription("Shortcut for `werf images cleanup && werf stages cleanup`"),
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfGitTagsExpiryDatePeriodPolicy, common.WerfGitTagsLimitPolicy, common.WerfGitCommitsExpiryDatePeriodPolicy, common.WerfGitCommitsLimitPolicy, common.WerfCleanupRegistryPassword, common.WerfDockerConfig, common.WerfIgnoreCIDockerAutologin, common.WerfInsecureRegistry, common.WerfHome),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			common.LogVersion()

			return common.LogRunningTime(func() error {
				err := runCleanup()
				if err != nil {
					return fmt.Errorf("cleanup failed: %s", err)
				}

				return nil
			})
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)
	common.SetupStagesRepo(&CommonCmdData, cmd)
	common.SetupImagesRepo(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.RegistryUsername, "registry-username", "", "", "Docker registry username (granted read-write permission)")
	cmd.Flags().StringVarP(&CmdData.RegistryPassword, "registry-password", "", "", "Docker registry password (granted read-write permission)")

	cmd.Flags().BoolVarP(&CmdData.WithoutKube, "without-kube", "", false, "Do not skip deployed kubernetes images")

	cmd.Flags().BoolVarP(&CmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runCleanup() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	kube.Init(kube.InitOptions{})

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

	projectName := werfConfig.Meta.Project

	imagesRepo, err := common.GetImagesRepo(projectName, &CommonCmdData)
	if err != nil {
		return err
	}

	stagesRepo, err := common.GetStagesRepo(&CommonCmdData)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetCleanupDockerAuthorizer(projectTmpDir, CmdData.RegistryUsername, CmdData.RegistryPassword, imagesRepo)
	if err != nil {
		return err
	}

	if err := dockerAuthorizer.Login(imagesRepo); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	var imagesNames []string
	for _, image := range werfConfig.Images {
		imagesNames = append(imagesNames, image.Name)
	}

	commonRepoOptions := cleanup.CommonRepoOptions{
		ImagesRepo:  imagesRepo,
		StagesRepo:  stagesRepo,
		ImagesNames: imagesNames,
		DryRun:      CmdData.DryRun,
	}

	var localGitRepo *git_repo.Local
	gitDir := path.Join(projectDir, ".git")
	if exist, err := util.DirExists(gitDir); err != nil {
		return err
	} else if exist {
		localGitRepo = &git_repo.Local{
			Path:   projectDir,
			GitDir: gitDir,
		}
	}

	commonProjectOptions := cleanup.CommonProjectOptions{
		ProjectName:   projectName,
		CommonOptions: cleanup.CommonOptions{DryRun: CmdData.DryRun},
	}

	imagesCleanupOptions := cleanup.ImagesCleanupOptions{
		CommonRepoOptions: commonRepoOptions,
		LocalGit:          localGitRepo,
		WithoutKube:       CmdData.WithoutKube,
	}

	stagesCleanupOptions := cleanup.StagesCleanupOptions{
		CommonRepoOptions:    commonRepoOptions,
		CommonProjectOptions: commonProjectOptions,
	}

	cleanupOptions := cleanup.CleanupOptions{
		StagesCleanupOptions: stagesCleanupOptions,
		ImagesCleanupOptions: imagesCleanupOptions,
	}

	if err := cleanup.Cleanup(cleanupOptions); err != nil {
		return err
	}

	return nil
}
