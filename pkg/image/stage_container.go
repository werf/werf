package image

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/docker/cli/cli/command"
	commandContainer "github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/util"
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

func NewStageImageContainer(image *Stage) *StageContainer {
	c := &StageContainer{}
	c.Image = image
	c.Name = fmt.Sprintf("dapp.build.%v", util.GenerateConsistentRandomString(10))
	c.RunOptions = NewStageContainerOptions()
	c.CommitChangeOptions = NewStageContainerOptions()
	c.ServiceCommitChangeOptions = NewStageContainerOptions()
	return c
}

func (c *StageContainer) runArgs(cli *command.DockerCli, apiClient *client.Client) ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.Name))

	runOptions, err := c.runOptions(cli, apiClient)
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.Image.FromImage.MustGetId(apiClient)
	if err != nil {
		return nil, err
	}

	args = append(args, runArgs...)
	args = append(args, fromImageId)
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

func (c *StageContainer) introspectBeforeArgs(cli *command.DockerCli, apiClient *client.Client) ([]string, error) {
	args, err := c.introspectArgsBase(cli, apiClient)
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.Image.FromImage.MustGetId(apiClient)
	if err != nil {
		return nil, err
	}

	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageContainer) introspectArgs(cli *command.DockerCli, apiClient *client.Client) ([]string, error) {
	args, err := c.introspectArgsBase(cli, apiClient)
	if err != nil {
		return nil, err
	}

	imageId, err := c.Image.MustGetId(apiClient)
	if err != nil {
		return nil, err
	}

	args = append(args, imageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageContainer) introspectArgsBase(cli *command.DockerCli, apiClient *client.Client) ([]string, error) {
	var args []string

	runOptions, err := c.introspectOptions(cli, apiClient)
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	args = append(args, []string{"-ti", "--rm"}...)
	args = append(args, runArgs...)

	return args, nil
}

func (c *StageContainer) runOptions(cli *command.DockerCli, apiClient *client.Client) (*StageContainerOptions, error) {
	serviceRunOptions, err := c.ServiceRunOptions(cli, apiClient)
	if err != nil {
		return nil, err
	}
	return serviceRunOptions.merge(c.RunOptions), nil
}

func (c *StageContainer) ServiceRunOptions(cli *command.DockerCli, apiClient *client.Client) (*StageContainerOptions, error) {
	serviceRunOptions := NewStageContainerOptions()
	serviceRunOptions.Workdir = "/"
	serviceRunOptions.Entrypoint = []string{dappdeps.BaseBinPath("bash")}
	serviceRunOptions.User = "0:0"

	baseContainerName, err := dappdeps.BaseContainer(cli, apiClient)
	if err != nil {
		return nil, err
	}

	toolchainContainerName, err := dappdeps.ToolchainContainer(cli, apiClient)
	if err != nil {
		return nil, err
	}
	serviceRunOptions.VolumesFrom = []string{baseContainerName, toolchainContainerName}

	return serviceRunOptions, nil
}

func (c *StageContainer) introspectOptions(cli *command.DockerCli, apiClient *client.Client) (*StageContainerOptions, error) {
	return c.runOptions(cli, apiClient)
}

func (c *StageContainer) commitChanges(apiClient *client.Client) ([]string, error) {
	commitOptions, err := c.commitOptions(apiClient)
	if err != nil {
		return nil, err
	}

	commitChanges, err := commitOptions.toCommitChanges(apiClient)
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *StageContainer) commitOptions(apiClient *client.Client) (*StageContainerOptions, error) {
	inheritedCommitOptions, err := c.inheritedCommitOptions(apiClient)
	if err != nil {
		return nil, err
	}

	commitOptions := inheritedCommitOptions.merge(c.ServiceCommitChangeOptions.merge(c.CommitChangeOptions))
	return commitOptions, nil
}

func (c *StageContainer) inheritedCommitOptions(apiClient *client.Client) (*StageContainerOptions, error) {
	inheritedOptions := NewStageContainerOptions()

	if c.Image.FromImage == nil {
		panic(fmt.Sprintf("runtime error: FromImage should be (%s)", c.Image.Name))
	}

	fromImageInspect, err := c.Image.FromImage.MustGetInspect(apiClient)
	if err != nil {
		return nil, err
	}

	inheritedOptions.Entrypoint = fromImageInspect.Config.Entrypoint
	inheritedOptions.Cmd = fromImageInspect.Config.Cmd
	if fromImageInspect.Config.WorkingDir != "" {
		inheritedOptions.Workdir = fromImageInspect.Config.WorkingDir
	} else {
		inheritedOptions.Workdir = "/"
	}
	return inheritedOptions, nil
}

func (c *StageContainer) Run(cli *command.DockerCli, apiClient *client.Client) error {
	runArgs, err := c.runArgs(cli, apiClient)
	if err != nil {
		return err
	}

	if err := c.run(runArgs, cli); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) Introspect(cli *command.DockerCli, apiClient *client.Client) error {
	runArgs, err := c.introspectArgs(cli, apiClient)
	if err != nil {
		return err
	}

	if err := c.run(runArgs, cli); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) IntrospectBefore(cli *command.DockerCli, apiClient *client.Client) error {
	runArgs, err := c.introspectBeforeArgs(cli, apiClient)
	if err != nil {
		return err
	}

	if err := c.run(runArgs, cli); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) run(args []string, cli *command.DockerCli) error {
	cmd := commandContainer.NewRunCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}

func (c *StageContainer) Commit(apiClient *client.Client) (string, error) {
	commitChanges, err := c.commitChanges(apiClient)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	commitOptions := types.ContainerCommitOptions{Changes: commitChanges}
	response, err := apiClient.ContainerCommit(ctx, c.Name, commitOptions)
	if err != nil {
		return "", err
	}
	return response.ID, nil
}

func (c *StageContainer) Rm(apiClient *client.Client) error {
	err := apiClient.ContainerRemove(context.Background(), c.Name, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}
