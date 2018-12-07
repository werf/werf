package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
)

var syncCmdData struct {
	Repo             string
	RegistryUsername string
	RegistryPassword string

	DryRun bool
}

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Remove local stages cache for the images, that don't exist into the docker registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync()
		},
	}

	cmd.PersistentFlags().StringVarP(&syncCmdData.Repo, "repo", "", "", "Docker repository name to get images information")
	cmd.PersistentFlags().StringVarP(&syncCmdData.RegistryUsername, "registry-username", "", "", "Docker registry username (granted read permission)")
	cmd.PersistentFlags().StringVarP(&syncCmdData.RegistryPassword, "registry-password", "", "", "Docker registry password (granted read permission)")

	cmd.PersistentFlags().BoolVarP(&syncCmdData.DryRun, "dry-run", "", false, "Indicate what the command would do without actually doing that")

	return cmd
}

func runSync() error {
	if err := lock.Init(); err != nil {
		return err
	}

	projectDir, err := getProjectDir()
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	projectTmpDir, err := getTmpDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}

	projectName, err := getProjectName(projectDir)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	repoName, err := getRequiredRepoName(projectName, cleanupCmdData.Repo)
	if err != nil {
		return err
	}

	dockerAuthorizer, err := docker_authorizer.GetSyncDockerAuthorizer(projectTmpDir, syncCmdData.RegistryUsername, syncCmdData.RegistryPassword, repoName)
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

	commonProjectOptions := cleanup.CommonProjectOptions{
		ProjectName:   projectName,
		CommonOptions: cleanup.CommonOptions{DryRun: syncCmdData.DryRun},
	}

	commonRepoOptions := cleanup.CommonRepoOptions{
		Repository: repoName,
		DimgsNames: dimgNames,
		DryRun:     syncCmdData.DryRun,
	}

	if err := cleanup.ProjectDimgstagesSync(commonProjectOptions, commonRepoOptions); err != nil {
		return err
	}

	return nil
}
