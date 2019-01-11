package cleanup

import (
	"fmt"
	"path"

	"github.com/flant/dapp/cmd/dapp/common"
	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/project_tmp_dir"
	"github.com/flant/kubedog/pkg/kube"

	"github.com/flant/dapp/pkg/util"
	"github.com/spf13/cobra"
)

var CmdData struct {
	Repo             string
	RegistryUsername string
	RegistryPassword string

	WithoutKube bool

	DryRun bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup project images in docker registry by policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runCleanup()
			if err != nil {
				return fmt.Errorf("cleanup failed: %s", err)
			}
			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.PersistentFlags().StringVarP(&CmdData.Repo, "repo", "", "", "Docker repository name")
	cmd.PersistentFlags().StringVarP(&CmdData.RegistryUsername, "registry-username", "", "", "Docker registry username (granted read-write permission)")
	cmd.PersistentFlags().StringVarP(&CmdData.RegistryPassword, "registry-password", "", "", "Docker registry password (granted read-write permission)")

	cmd.PersistentFlags().BoolVarP(&CmdData.WithoutKube, "without-kube", "", false, "Do not skip deployed kubernetes images")

	cmd.PersistentFlags().BoolVarP(&CmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runCleanup() error {
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
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

	projectTmpDir, err := project_tmp_dir.Get()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer project_tmp_dir.Release(projectTmpDir)

	dappfile, err := common.GetDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	projectName := dappfile.Meta.Project

	repoName, err := common.GetRequiredRepoName(projectName, CmdData.Repo)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetCleanupDockerAuthorizer(projectTmpDir, CmdData.RegistryUsername, CmdData.RegistryPassword, repoName)
	if err != nil {
		return err
	}

	if err := dockerAuthorizer.Login(repoName); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	var dimgNames []string
	for _, dimg := range dappfile.Dimgs {
		dimgNames = append(dimgNames, dimg.Name)
	}

	commonRepoOptions := cleanup.CommonRepoOptions{
		Repository: repoName,
		DimgsNames: dimgNames,
		DryRun:     CmdData.DryRun,
	}

	var localRepo *git_repo.Local
	gitDir := path.Join(projectDir, ".git")
	if exist, err := util.DirExists(gitDir); err != nil {
		return err
	} else if exist {
		localRepo = &git_repo.Local{
			Path:   projectDir,
			GitDir: gitDir,
		}
	}

	cleanupOptions := cleanup.CleanupOptions{
		CommonRepoOptions: commonRepoOptions,
		LocalRepo:         localRepo,
		WithoutKube:       CmdData.WithoutKube,
	}

	if err := cleanup.Cleanup(cleanupOptions); err != nil {
		return err
	}

	return nil
}
