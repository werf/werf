package flush

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/docker_authorizer"
	"github.com/flant/werf/pkg/cleanup"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/project_tmp_dir"
	"github.com/flant/werf/pkg/werf"
)

var CmdData struct {
	Repo             string
	RegistryUsername string
	RegistryPassword string

	WithImages bool

	DryRun bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "flush",
		DisableFlagsInUseLine: true,
		Short:                 "Delete project images in local docker storage and specified docker registry",
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(common.WerfInsecureRegistry, common.WerfHome),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runFlush()
			if err != nil {
				return fmt.Errorf("flush failed: %s", err)
			}
			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.Flags().StringVarP(&CmdData.Repo, "repo", "", "", "Docker repository name")
	cmd.Flags().StringVarP(&CmdData.RegistryUsername, "registry-username", "", "", "Docker registry username (granted read-write permission)")
	cmd.Flags().StringVarP(&CmdData.RegistryPassword, "registry-password", "", "", "Docker registry password (granted read-write permission)")

	cmd.Flags().BoolVarP(&CmdData.WithImages, "with-images", "", false, "Delete images (not only stages cache)")

	cmd.Flags().BoolVarP(&CmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runFlush() error {
	if err := werf.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
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

	repoName := common.GetOptionalRepoName(projectName, CmdData.Repo)

	if repoName != "" {
		if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
			return err
		}

		projectTmpDir, err := project_tmp_dir.Get()
		if err != nil {
			return fmt.Errorf("getting project tmp dir failed: %s", err)
		}
		defer project_tmp_dir.Release(projectTmpDir)

		dockerAuthorizer, err := docker_authorizer.GetFlushDockerAuthorizer(projectTmpDir, CmdData.RegistryUsername, CmdData.RegistryPassword)
		if err != nil {
			return err
		}

		if err := dockerAuthorizer.Login(repoName); err != nil {
			return err
		}

		var imageNames []string
		for _, image := range werfConfig.Images {
			imageNames = append(imageNames, image.Name)
		}

		commonRepoOptions := cleanup.CommonRepoOptions{
			Repository:  repoName,
			ImagesNames: imageNames,
			DryRun:      CmdData.DryRun,
		}

		if err := cleanup.RepoImagesFlush(CmdData.WithImages, commonRepoOptions); err != nil {
			return err
		}
	} else {
		if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
			return err
		}
	}

	commonProjectOptions := cleanup.CommonProjectOptions{
		ProjectName:   projectName,
		CommonOptions: cleanup.CommonOptions{DryRun: CmdData.DryRun},
	}

	if err := cleanup.ProjectImagesFlush(CmdData.WithImages, commonProjectOptions); err != nil {
		return err
	}

	return nil
}
