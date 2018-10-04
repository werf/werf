package image

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/docker"
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

func (c *StageContainer) AddRunCommands(commands []string) {
	c.RunCommands = append(c.RunCommands, commands...)
}

func (c *StageContainer) AddServiceRunCommands(commands []string) {
	c.ServiceRunCommands = append(c.ServiceRunCommands, commands...)
}

func (c *StageContainer) runArgs() ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.Name))

	runOptions, err := c.runOptions()
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.Image.FromImage.MustGetId()
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

func (c *StageContainer) introspectBeforeArgs() ([]string, error) {
	args, err := c.introspectArgsBase()
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.Image.FromImage.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageContainer) introspectArgs() ([]string, error) {
	args, err := c.introspectArgsBase()
	if err != nil {
		return nil, err
	}

	imageId, err := c.Image.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, imageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageContainer) introspectArgsBase() ([]string, error) {
	var args []string

	runOptions, err := c.introspectOptions()
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

func (c *StageContainer) runOptions() (*StageContainerOptions, error) {
	serviceRunOptions, err := c.ServiceRunOptions()
	if err != nil {
		return nil, err
	}
	return serviceRunOptions.merge(c.RunOptions), nil
}

func (c *StageContainer) ServiceRunOptions() (*StageContainerOptions, error) {
	serviceRunOptions := NewStageContainerOptions()
	serviceRunOptions.Workdir = "/"
	serviceRunOptions.Entrypoint = []string{dappdeps.BaseBinPath("bash")}
	serviceRunOptions.User = "0:0"

	baseContainerName, err := dappdeps.BaseContainer()
	if err != nil {
		return nil, err
	}

	toolchainContainerName, err := dappdeps.ToolchainContainer()
	if err != nil {
		return nil, err
	}
	serviceRunOptions.VolumesFrom = []string{baseContainerName, toolchainContainerName}

	return serviceRunOptions, nil
}

func (c *StageContainer) introspectOptions() (*StageContainerOptions, error) {
	return c.runOptions()
}

func (c *StageContainer) commitChanges() ([]string, error) {
	commitOptions, err := c.commitOptions()
	if err != nil {
		return nil, err
	}

	commitChanges, err := commitOptions.toCommitChanges()
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *StageContainer) commitOptions() (*StageContainerOptions, error) {
	inheritedCommitOptions, err := c.inheritedCommitOptions()
	if err != nil {
		return nil, err
	}

	commitOptions := inheritedCommitOptions.merge(c.ServiceCommitChangeOptions.merge(c.CommitChangeOptions))
	return commitOptions, nil
}

func (c *StageContainer) inheritedCommitOptions() (*StageContainerOptions, error) {
	inheritedOptions := NewStageContainerOptions()

	if c.Image.FromImage == nil {
		panic(fmt.Sprintf("runtime error: FromImage should be (%s)", c.Image.Name))
	}

	fromImageInspect, err := c.Image.FromImage.MustGetInspect()
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

func (c *StageContainer) Run() error {
	runArgs, err := c.runArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}

func (c *StageContainer) Introspect() error {
	runArgs, err := c.introspectArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) IntrospectBefore() error {
	runArgs, err := c.introspectBeforeArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) Commit() (string, error) {
	commitChanges, err := c.commitChanges()
	if err != nil {
		return "", err
	}

	commitOptions := types.ContainerCommitOptions{Changes: commitChanges}
	id, err := docker.ContainerCommit(c.Name, commitOptions)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *StageContainer) Rm() error {
	return docker.ContainerRemove(c.Name, types.ContainerRemoveOptions{})
}
