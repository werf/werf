package main

import (
	"fmt"
	"os"
	"time"

	"github.com/flant/dapp/cmd/dapp/docker_authorizer"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ssh_agent"
	"github.com/flant/dapp/pkg/true_git"
	"github.com/flant/kubedog/pkg/kube"
	"github.com/spf13/cobra"
)

var deployCmdData struct {
	HelmReleaseName string

	Namespace   string
	KubeContext string
	Timeout     int

	Values       []string
	SecretValues []string
	Set          []string

	Repo             string
	RegistryUsername string
	RegistryPassword string
	WithoutRegistry  bool

	Tag        []string
	TagBranch  bool
	TagBuildID bool
	TagCI      bool
	TagCommit  bool
}

func newDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "deploy HELM_RELEASE_NAME",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deployCmdData.HelmReleaseName = args[0]

			err := runDeploy()
			if err != nil {
				return fmt.Errorf("deploy failed: %s", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&deployCmdData.Namespace, "namespace", "", "", "Kubernetes namespace")
	cmd.PersistentFlags().StringVarP(&deployCmdData.KubeContext, "kube-context", "", "", "Kubernetes config context")
	cmd.PersistentFlags().IntVarP(&deployCmdData.Timeout, "timeout", "t", 0, "watch timeout in seconds")

	cmd.PersistentFlags().StringArrayVarP(&deployCmdData.Values, "values", "", []string{}, "Additional helm values")
	cmd.PersistentFlags().StringArrayVarP(&deployCmdData.SecretValues, "secret-values", "", []string{}, "Additional helm secret values")
	cmd.PersistentFlags().StringArrayVarP(&deployCmdData.Set, "set", "", []string{}, "Additional helm sets")

	cmd.PersistentFlags().StringVarP(&deployCmdData.Repo, "repo", "", "", "Docker repository name to get images ids from. CI_REGISTRY_IMAGE will be used by default if available.")
	cmd.PersistentFlags().StringVarP(&deployCmdData.RegistryUsername, "registry-username", "", "", "Docker registry username")
	cmd.PersistentFlags().StringVarP(&deployCmdData.RegistryPassword, "registry-password", "", "", "Docker registry password")
	cmd.PersistentFlags().BoolVarP(&deployCmdData.WithoutRegistry, "without-registry", "", false, "Do not get images info from registry")

	cmd.PersistentFlags().StringArrayVarP(&pushCmdData.Tag, "tag", "", []string{}, "Add tag (can be used one or more times)")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagBranch, "tag-branch", "", false, "Tag by git branch")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagBuildID, "tag-build-id", "", false, "Tag by CI build id")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagCI, "tag-ci", "", false, "Tag by CI branch and tag")
	cmd.PersistentFlags().BoolVarP(&pushCmdData.TagCommit, "tag-commit", "", false, "Tag by git commit")

	return cmd
}

func runDeploy() error {
	if err := dapp.Init(rootCmdData.TmpDir, rootCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := lock.Init(); err != nil {
		return err
	}

	if err := true_git.Init(); err != nil {
		return err
	}

	projectDir, err := getProjectDir()
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	projectName, err := getProjectName(projectDir)
	if err != nil {
		return fmt.Errorf("getting project name failed: %s", err)
	}

	projectTmpDir, err := getTmpDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}

	dappfile, err := parseDappfile(projectDir)
	if err != nil {
		return fmt.Errorf("dappfile parsing failed: %s", err)
	}

	var repo string
	if !deployCmdData.WithoutRegistry {
		var err error
		repo, err = getRequiredRepoName(projectName, deployCmdData.Repo)
		if err != nil {
			return err
		}
	}

	dockerAuthorizer, err := docker_authorizer.GetDeployDockerAuthorizer(projectTmpDir, deployCmdData.RegistryUsername, deployCmdData.RegistryPassword, repo)
	if err != nil {
		return err
	}

	if err := docker.Init(dockerAuthorizer.HostDockerConfigDir); err != nil {
		return err
	}

	if err := ssh_agent.Init(rootCmdData.SSHKeys); err != nil {
		return fmt.Errorf("cannot initialize ssh-agent: %s", err)
	}

	tag, err := getDeployTag(projectDir, deployCmdData.Tag, deployCmdData.TagBranch, deployCmdData.TagCommit, deployCmdData.TagBuildID, deployCmdData.TagCI)
	if err != nil {
		return err
	}

	kubeContext := os.Getenv("KUBECONTEXT")
	if kubeContext == "" {
		kubeContext = deployCmdData.KubeContext
	}
	err = kube.Init(kube.InitOptions{KubeContext: kubeContext})
	if err != nil {
		return fmt.Errorf("cannot initialize kube: %s", err)
	}

	namespace := getNamespace(deployCmdData.Namespace)

	return deploy.RunDeploy(projectName, projectDir, deployCmdData.HelmReleaseName, namespace, kubeContext, repo, tag, dappfile, deploy.DeployOptions{
		Values:          deployCmdData.Values,
		SecretValues:    deployCmdData.SecretValues,
		Set:             deployCmdData.Set,
		Timeout:         time.Duration(deployCmdData.Timeout) * time.Second,
		WithoutRegistry: deployCmdData.WithoutRegistry,
	})
}
