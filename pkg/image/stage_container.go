package image

import (
	"encoding/base64"
	"fmt"
	"strings"

	dockerClient "github.com/docker/cli/cli/command"
	dockerClientContainer "github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/flant/dapp/pkg/util"
	"golang.org/x/net/context"

	"github.com/flant/dapp/pkg/dappdeps"
)

type StageContainer struct {
	Image                      *Stage
	Name                       string
	RunCommands                []string
	ServiceRunCommands         []string
	RunOptions                 *StageContainerOptions
	CommitChangeOptions        *StageContainerOptions
	ServiceCommitChangeOptions *StageContainerOptions
}

func NewStageImageContainer() *StageContainer {
	c := &StageContainer{}
	token, err := util.GenerateConsistentRandomString(10)
	if err != nil {
		panic(err.Error())
	}
	c.Name = fmt.Sprintf("dapp.build.%v", token)
	c.RunOptions = NewStageContainerOptions()
	c.CommitChangeOptions = NewStageContainerOptions()
	c.ServiceCommitChangeOptions = NewStageContainerOptions()
	return c
}

func (c *StageContainer) runArgs(dockerClient *dockerClient.DockerCli, dockerApiClient *client.Client) ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.Name))

	runArgs, err := c.runOptions(dockerClient, dockerApiClient).toRunArgs()
	if err != nil {
		return nil, err
	}

	args = append(args, runArgs...)
	args = append(args, c.Image.From.BuiltId)
	args = append(args, "-ec")
	args = append(args, c.PreparedRunCommand())

	return args, nil
}

func (c *StageContainer) PreparedRunCommand() string {
	return ShelloutPack(strings.Join(c.PreparedRunCommands(), " && "))
}

func (c *StageContainer) PreparedRunCommands() []string {
	runCommands := append(c.RunCommands, c.ServiceRunCommands...)
	if len(runCommands) != 0 {
		return runCommands
	} else {
		return []string{dappdeps.BaseBinPath("true")}
	}
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | %s --decode)", base64.StdEncoding.EncodeToString([]byte(command)), dappdeps.BaseBinPath("base64"))
}

func (c *StageContainer) introspectArgs(dockerClient *dockerClient.DockerCli, dockerApiClient *client.Client) ([]string, error) {
	var args []string

	runArgs, err := c.introspectOptions(dockerClient, dockerApiClient).toRunArgs()
	if err != nil {
		return nil, err
	}

	args = append(args, runArgs...)
	args = append(args, []string{"-ti", "--rm"}...)
	args = append(args, c.Image.BuiltId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageContainer) runOptions(dockerClient *dockerClient.DockerCli, dockerApiClient *client.Client) *StageContainerOptions {
	return c.ServiceRunOptions(dockerClient, dockerApiClient).merge(c.RunOptions)
}

func (c *StageContainer) ServiceRunOptions(dockerClient *dockerClient.DockerCli, dockerApiClient *client.Client) *StageContainerOptions {
	serviceRunOptions := NewStageContainerOptions()
	serviceRunOptions.Workdir = "/"
	serviceRunOptions.Entrypoint = []string{dappdeps.BaseBinPath("bash")}
	serviceRunOptions.User = "0:0"
	serviceRunOptions.VolumesFrom = []string{dappdeps.BaseContainer(dockerClient, dockerApiClient), dappdeps.ToolchainContainer(dockerClient, dockerApiClient)}
	return serviceRunOptions
}

func (c *StageContainer) introspectOptions(dockerClient *dockerClient.DockerCli, dockerApiClient *client.Client) *StageContainerOptions {
	return c.runOptions(dockerClient, dockerApiClient)
}

func (c *StageContainer) commitChanges(client *client.Client) ([]string, error) {
	commitChanges, err := c.commitOptions(client).toCommitChanges(client)
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *StageContainer) commitOptions(client *client.Client) *StageContainerOptions {
	return c.inheritedCommitOptions(client).merge(c.ServiceCommitChangeOptions.merge(c.CommitChangeOptions))
}

func (c *StageContainer) inheritedCommitOptions(client *client.Client) *StageContainerOptions {
	inheritedOptions := NewStageContainerOptions()
	if c.Image.From != nil {
		var inspect *types.ImageInspect
		if c.Image.From.BuiltInspect != nil {
			inspect = c.Image.From.BuiltInspect
		} else {
			inspect = c.Image.From.Inspect
		}

		inheritedOptions.Entrypoint = inspect.Config.Entrypoint
		inheritedOptions.Cmd = inspect.Config.Cmd
		if inspect.Config.WorkingDir != "" {
			inheritedOptions.Workdir = inspect.Config.WorkingDir
		} else {
			inheritedOptions.Workdir = "/"
		}
	}
	return inheritedOptions
}

func (c *StageContainer) Run(dockerClient *dockerClient.DockerCli, dockerApiClient *client.Client) error {
	runArgs, err := c.runArgs(dockerClient, dockerApiClient)
	if err != nil {
		return err
	}

	if err := c.run(runArgs, dockerClient); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) Introspect(dockerClient *dockerClient.DockerCli, dockerApiClient *client.Client) error {
	runArgs, err := c.introspectArgs(dockerClient, dockerApiClient)
	if err != nil {
		return err
	}

	if err := c.run(runArgs, dockerClient); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) run(args []string, dockerCli *dockerClient.DockerCli) error {
	command := dockerClientContainer.NewRunCommand(dockerCli)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(args)

	err := command.Execute()
	if err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}

func (c *StageContainer) Commit(client *client.Client) (string, error) {
	commitChanges, err := c.commitChanges(client)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	commitOptions := types.ContainerCommitOptions{Changes: commitChanges}
	response, err := client.ContainerCommit(ctx, c.Name, commitOptions)
	if err != nil {
		return "", fmt.Errorf("container commit failed: %s", err.Error())
	}
	return response.ID, nil
}

func (c *StageContainer) Rm(dockerApiClient *client.Client) error {
	err := dockerApiClient.ContainerRemove(context.Background(), c.Name, types.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("container rm failed: %s", err.Error())
	}
	return nil
}
