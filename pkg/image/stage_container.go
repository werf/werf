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
	image                      *Stage
	name                       string
	runCommands                []string
	serviceRunCommands         []string
	runOptions                 *StageContainerOptions
	commitChangeOptions        *StageContainerOptions
	serviceCommitChangeOptions *StageContainerOptions
}

func newStageImageContainer(image *Stage) *StageContainer {
	c := &StageContainer{}
	c.image = image
	c.name = fmt.Sprintf("dapp.build.%v", util.GenerateConsistentRandomString(10))
	c.runOptions = newStageContainerOptions()
	c.commitChangeOptions = newStageContainerOptions()
	c.serviceCommitChangeOptions = newStageContainerOptions()
	return c
}

func (c *StageContainer) Name() string {
	return c.name
}

func (c *StageContainer) AddRunCommands(commands ...string) {
	c.runCommands = append(c.runCommands, commands...)
}

func (c *StageContainer) AddServiceRunCommands(commands ...string) {
	c.serviceRunCommands = append(c.serviceRunCommands, commands...)
}

func (c *StageContainer) RunOptions() *StageContainerOptions {
	return c.runOptions
}

func (c *StageContainer) CommitChangeOptions() *StageContainerOptions {
	return c.commitChangeOptions
}

func (c *StageContainer) ServiceCommitChangeOptions() *StageContainerOptions {
	return c.serviceCommitChangeOptions
}

func (c *StageContainer) prepareRunArgs() ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.name))

	runOptions, err := c.prepareRunOptions()
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.image.fromImage.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, runArgs...)
	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, c.prepareRunCommand())

	return args, nil
}

func (c *StageContainer) prepareRunCommand() string {
	return shelloutPack(strings.Join(c.prepareRunCommands(), " && "))
}

func (c *StageContainer) prepareRunCommands() []string {
	runCommands := c.prepareAllRunCommands()
	if len(runCommands) != 0 {
		return runCommands
	} else {
		return []string{dappdeps.BaseBinPath("true")}
	}
}

func (c *StageContainer) prepareAllRunCommands() []string {
	return append(c.runCommands, c.serviceRunCommands...)
}

func shelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | %s --decode)", base64.StdEncoding.EncodeToString([]byte(command)), dappdeps.BaseBinPath("base64"))
}

func (c *StageContainer) prepareIntrospectBeforeArgs() ([]string, error) {
	args, err := c.prepareIntrospectArgsBase()
	if err != nil {
		return nil, err
	}

	fromImageId, err := c.image.fromImage.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageContainer) prepareIntrospectArgs() ([]string, error) {
	args, err := c.prepareIntrospectArgsBase()
	if err != nil {
		return nil, err
	}

	imageId, err := c.image.MustGetId()
	if err != nil {
		return nil, err
	}

	args = append(args, imageId)
	args = append(args, "-ec")
	args = append(args, dappdeps.BaseBinPath("bash"))

	return args, nil
}

func (c *StageContainer) prepareIntrospectArgsBase() ([]string, error) {
	var args []string

	runOptions, err := c.prepareIntrospectOptions()
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

func (c *StageContainer) prepareRunOptions() (*StageContainerOptions, error) {
	serviceRunOptions, err := c.prepareServiceRunOptions()
	if err != nil {
		return nil, err
	}
	return serviceRunOptions.merge(c.runOptions), nil
}

func (c *StageContainer) prepareServiceRunOptions() (*StageContainerOptions, error) {
	serviceRunOptions := newStageContainerOptions()
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

func (c *StageContainer) prepareIntrospectOptions() (*StageContainerOptions, error) {
	return c.prepareRunOptions()
}

func (c *StageContainer) prepareCommitChanges() ([]string, error) {
	commitOptions, err := c.prepareCommitOptions()
	if err != nil {
		return nil, err
	}

	commitChanges, err := commitOptions.toCommitChanges()
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *StageContainer) prepareCommitOptions() (*StageContainerOptions, error) {
	inheritedCommitOptions, err := c.prepareInheritedCommitOptions()
	if err != nil {
		return nil, err
	}

	commitOptions := inheritedCommitOptions.merge(c.serviceCommitChangeOptions.merge(c.commitChangeOptions))
	return commitOptions, nil
}

func (c *StageContainer) prepareInheritedCommitOptions() (*StageContainerOptions, error) {
	inheritedOptions := newStageContainerOptions()

	if c.image.fromImage == nil {
		panic(fmt.Sprintf("runtime error: FromImage should be (%s)", c.image.name))
	}

	fromImageInspect, err := c.image.fromImage.MustGetInspect()
	if err != nil {
		return nil, err
	}

	inheritedOptions.Entrypoint = fromImageInspect.Config.Entrypoint
	inheritedOptions.Cmd = fromImageInspect.Config.Cmd
	inheritedOptions.User = fromImageInspect.Config.User
	if fromImageInspect.Config.WorkingDir != "" {
		inheritedOptions.Workdir = fromImageInspect.Config.WorkingDir
	} else {
		inheritedOptions.Workdir = "/"
	}
	return inheritedOptions, nil
}

func (c *StageContainer) run() error {
	runArgs, err := c.prepareRunArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}

func (c *StageContainer) introspect() error {
	runArgs, err := c.prepareIntrospectArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) introspectBefore() error {
	runArgs, err := c.prepareIntrospectBeforeArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun(runArgs...); err != nil {
		return err
	}

	return nil
}

func (c *StageContainer) commit() (string, error) {
	commitChanges, err := c.prepareCommitChanges()
	if err != nil {
		return "", err
	}

	commitOptions := types.ContainerCommitOptions{Changes: commitChanges}
	id, err := docker.ContainerCommit(c.name, commitOptions)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *StageContainer) rm() error {
	return docker.ContainerRemove(c.name, types.ContainerRemoveOptions{})
}
