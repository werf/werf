package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/git_repo"
	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/dapp/pkg/slug"
	"github.com/spf13/cobra"
)

var rootCmdData struct {
	Name    string
	Dir     string
	TmpDir  string
	HomeDir string
	SSHKeys []string
}

func main() {
	cmd := &cobra.Command{
		Use:          "dapp",
		SilenceUsage: true,
	}

	cmd.AddCommand(
		newBuildCmd(),
		newPushCmd(),
		newBPCmd(),
	)

	cmd.PersistentFlags().StringVarP(&rootCmdData.Name, "name", "", "", `Use custom dapp name.
Chaging default name will cause full cache rebuild.
By default dapp name is the last element of remote.origin.url from project git,
or it is the name of the directory where Dappfile resides.`)
	cmd.PersistentFlags().StringVarP(&rootCmdData.Dir, "dir", "", "", "Change to the specified directory to find dappfile")
	cmd.PersistentFlags().StringVarP(&rootCmdData.TmpDir, "tmp-dir", "", "", "Use specified dir to store tmp files and dirs (use system tmp dir by default)")
	cmd.PersistentFlags().StringVarP(&rootCmdData.HomeDir, "home-dir", "", "", "Use specified dir to store dapp cache files and dirs (use ~/.dapp by default)")
	cmd.PersistentFlags().StringArrayVarP(&rootCmdData.SSHKeys, "ssh-key", "", []string{}, "Enable only specified ssh keys (use system ssh-agent by default)")

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getProjectName(projectDir string) (string, error) {
	name := path.Base(projectDir)

	if rootCmdData.Name != "" {
		name = rootCmdData.Name
	} else {
		exist, err := isGitOwnRepoExists(projectDir)
		if err != nil {
			return "", err
		}

		if exist {
			remoteOriginUrl, err := gitOwnRepoOriginUrl(projectDir)
			if err != nil {
				return "", err
			}

			if remoteOriginUrl != "" {
				parts := strings.Split(remoteOriginUrl, "/")
				repoName := parts[len(parts)-1]

				gitEnding := ".git"
				if strings.HasSuffix(repoName, gitEnding) {
					repoName = repoName[0 : len(repoName)-len(gitEnding)]
				}

				name = repoName
			}
		}
	}

	return slug.Slug(name), nil
}

func parseDappfile(projectDir string) ([]*config.Dimg, error) {
	for _, dappfileName := range []string{"dappfile.yml", "dappfile.yaml"} {
		dappfilePath := path.Join(projectDir, dappfileName)
		if exist, err := file.FileExists(dappfilePath); err != nil {
			return nil, err
		} else if exist {
			return config.ParseDimgs(dappfilePath)
		}
	}

	return nil, errors.New("dappfile.y[a]ml not found")
}

func getProjectDir() (string, error) {
	if rootCmdData.Dir != "" {
		return rootCmdData.Dir, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return currentDir, nil
}

func getProjectTmpDir() (string, error) {
	return ioutil.TempDir(dapp.GetTmpDir(), "dapp-")
}

func getProjectBuildDir(projectName string) (string, error) {
	projectBuildDir := path.Join(dapp.GetHomeDir(), "build", projectName)

	if err := os.MkdirAll(projectBuildDir, os.ModePerm); err != nil {
		return "", err
	}

	return projectBuildDir, nil
}

func isGitOwnRepoExists(projectDir string) (bool, error) {
	fileInfo, err := os.Stat(path.Join(projectDir, ".git"))
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	return fileInfo.IsDir(), nil
}

func gitOwnRepoOriginUrl(projectDir string) (string, error) {
	localGitRepo := &git_repo.Local{
		Path:   projectDir,
		GitDir: path.Join(projectDir, ".git"),
	}

	remoteOriginUrl, err := localGitRepo.RemoteOriginUrl()
	if err != nil {
		return "", nil
	}

	return remoteOriginUrl, nil
}

func getRequiredRepoName(projectName, repoOption string) (string, error) {
	res := getOptionalRepoName(projectName, repoOption)
	if res == "" {
		return "", fmt.Errorf("CI_REGISTRY_IMAGE variable or repo option required!")
	}
	return res, nil
}

func getOptionalRepoName(projectName, repoOption string) string {
	if repoOption == ":minikube" {
		return fmt.Sprintf("localhost:5000/%s", projectName)
	} else if repoOption != "" {
		return repoOption
	}

	ciRegistryImage := os.Getenv("CI_REGISTRY_IMAGE")
	if ciRegistryImage != "" {
		return ciRegistryImage
	}

	return ""
}
