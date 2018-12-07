package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/cleanup"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
)

var flushCmdData struct {
	Repo             string
	RegistryUsername string
	RegistryPassword string

	WithDimgs bool

	DryRun bool
}

func newFlushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flush",
		Short: "Delete project images in local docker storage and specified docker registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFlush()
		},
	}

	cmd.PersistentFlags().StringVarP(&flushCmdData.Repo, "repo", "", "", "docker repository name")
	cmd.PersistentFlags().StringVarP(&flushCmdData.RegistryUsername, "registry-username", "", "", "docker registry username (granted read-write permission)")
	cmd.PersistentFlags().StringVarP(&flushCmdData.RegistryPassword, "registry-password", "", "", "docker registry password (granted read-write permission)")
	cmd.PersistentFlags().BoolVarP(&flushCmdData.WithDimgs, "with-dimgs", "", false, "delete images (not only stages cache)")

	cmd.PersistentFlags().BoolVarP(&flushCmdData.DryRun, "dry-run", "", false, "indicate what the command would do without actually doing that")

	return cmd
}

func runFlush() error {
	if err := lock.Init(); err != nil {
		return err
	}

	projectDir, err := getProjectDir()
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	if flushCmdData.Repo != "" {
		projectTmpDir, err := getProjectTmpDir()
		if err != nil {
			return fmt.Errorf("getting project tmp dir failed: %s", err)
		}

		dockerAuthorizer, err := docker_authorizer.GetFlushDockerAuthorizer(projectTmpDir, flushCmdData.RegistryUsername, flushCmdData.RegistryPassword)
		if err != nil {
			return err
		}

		if err := dockerAuthorizer.Login(flushCmdData.Repo); err != nil {
			return err
		}
	}

	if err := docker.Init(docker_authorizer.GetHomeDockerConfigDir()); err != nil {
		return err
	}

	if flushCmdData.Repo != "" {
		dappfile, err := parseDappfile(projectDir)
		if err != nil {
			return fmt.Errorf("dappfile parsing failed: %s", err)
		}

		var dimgNames []string
		for _, dimg := range dappfile {
			dimgNames = append(dimgNames, dimg.Name)
		}

		commonRepoOptions := cleanup.CommonRepoOptions{
			Repository: flushCmdData.Repo,
			DimgsNames: dimgNames,
			DryRun:     flushCmdData.DryRun,
		}

		if err := cleanup.RepoImagesFlush(flushCmdData.WithDimgs, commonRepoOptions); err != nil {
			return err
		}
	}

	projectName, err := getProjectName(projectDir)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	commonProjectOptions := cleanup.CommonProjectOptions{
		ProjectName:   projectName,
		CommonOptions: cleanup.CommonOptions{DryRun: flushCmdData.DryRun},
	}

	if err := cleanup.ProjectImagesFlush(flushCmdData.WithDimgs, commonProjectOptions); err != nil {
		return err
	}

	return nil
}
