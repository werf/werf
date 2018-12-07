package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/kubedog/pkg/kube"
)

var cleanupCmdData struct {
	Repo             string
	RegistryUsername string
	RegistryPassword string

	WithoutKube bool

	DryRun bool
}

func newCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup project images in docker registry by policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanup()
		},
	}

	cmd.PersistentFlags().StringVarP(&cleanupCmdData.Repo, "repo", "", "", "docker repository name")
	cmd.PersistentFlags().StringVarP(&cleanupCmdData.RegistryUsername, "registry-username", "", "", "docker registry username (granted read-write permission)")
	cmd.PersistentFlags().StringVarP(&cleanupCmdData.RegistryPassword, "registry-password", "", "", "docker registry password (granted read-write permission)")

	cmd.PersistentFlags().BoolVarP(&cleanupCmdData.WithoutKube, "without-kube", "", false, "do not skip deployed kubernetes images")

	cmd.PersistentFlags().BoolVarP(&cleanupCmdData.DryRun, "dry-run", "", false, "indicate what the command would do without actually doing that")

	return cmd
}

func runCleanup() error {
	if err := lock.Init(); err != nil {
		return err
	}

	kube.Init(kube.InitOptions{})

	projectDir, err := getProjectDir()
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	projectName, err := getProjectName(projectDir)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	projectTmpDir, err := getProjectTmpDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}

	repoName, err := getRequiredRepoName(projectName, cleanupCmdData.Repo)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetCleanupDockerAuthorizer(projectTmpDir, cleanupCmdData.RegistryUsername, cleanupCmdData.RegistryPassword, repoName)
	if err != nil {
		return err
	}

	if err := dockerAuthorizer.Login(repoName); err != nil {
		return err
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	dappfile, err := parseDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	var dimgNames []string
	for _, dimg := range dappfile {
		dimgNames = append(dimgNames, dimg.Name)
	}

	commonRepoOptions := cleanup.CommonRepoOptions{
		Repository: repoName,
		DimgsNames: dimgNames,
		DryRun:     cleanupCmdData.DryRun,
	}

	localRepo := &git_repo.Local{}
	if exist, err := isGitOwnRepoExists(projectDir); err != nil {
		return err
	} else if exist {
		localRepo = localGitRepo(projectDir)
	}

	cleanupOptions := cleanup.CleanupOptions{
		CommonRepoOptions: commonRepoOptions,
		LocalRepo:         localRepo,
		WithoutKube:       cleanupCmdData.WithoutKube,
	}

	if err := cleanup.Cleanup(cleanupOptions); err != nil {
		return err
	}

	return nil
}
