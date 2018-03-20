package image

import (
	"fmt"
	dockerClient "github.com/docker/cli/cli/command"
	dockerClientContainer "github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type StageContainer struct {
	Image                      *Stage
	Name                       string
	RunCommands                []string
	ServiceRunCommands         []string
	PreparedRunCommand         string
	RunOptions                 *StageContainerOptions
	ServiceRunOptions          *StageContainerOptions
	CommitChangeOptions        *StageContainerOptions
	ServiceCommitChangeOptions *StageContainerOptions
}

func NewStageImageContainer() *StageContainer {
	c := &StageContainer{}
	//token, err := util.GenerateConsistentRandomString(10) // TODO Name generation (blocked by introspection)
	//if err != nil {
	//	panic(err.Error())
	//}
	//c.Name = fmt.Sprintf("dapp.%v", token)
	c.RunOptions = NewStageContainerOptions()
	c.ServiceRunOptions = NewStageContainerOptions()
	c.CommitChangeOptions = NewStageContainerOptions()
	c.ServiceCommitChangeOptions = NewStageContainerOptions()
	return c
}

func (c *StageContainer) runArgs() ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.Name))

	runArgs, err := c.runOptions().toRunArgs()
	if err != nil {
		return nil, err
	}

	args = append(args, runArgs...)
	args = append(args, c.Image.From.BuiltId)
	args = append(args, "-ec")
	args = append(args, c.PreparedRunCommand) // TODO using RunCommands && ServiceRunCommands (blocked by introspection)

	return args, nil
}

func (c *StageContainer) runOptions() *StageContainerOptions {
	return c.ServiceRunOptions.merge(c.RunOptions)
}

func (c *StageContainer) commitChanges(client *client.Client) ([]string, error) {
	commitChanges, err := c.commitOptions().toCommitChanges(client)
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *StageContainer) commitOptions() *StageContainerOptions {
	return c.ServiceCommitChangeOptions.merge(c.CommitChangeOptions)
}

func (c *StageContainer) Run(dockerCli *dockerClient.DockerCli) error {
	command := dockerClientContainer.NewRunCommand(dockerCli)

	runArgs, err := c.runArgs()
	if err != nil {
		return err
	}

	command.SetArgs(runArgs)
	command.SilenceUsage = true
	command.SilenceErrors = true

	err = command.Execute()
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
