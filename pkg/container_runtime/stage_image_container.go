package container_runtime

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/image"

	"github.com/docker/docker/api/types"

	"github.com/flant/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
)

type StageImageContainer struct {
	image                      *StageImage
	name                       string
	runCommands                []string
	serviceRunCommands         []string
	runOptions                 *StageImageContainerOptions
	commitChangeOptions        *StageImageContainerOptions
	serviceCommitChangeOptions *StageImageContainerOptions
}

func newStageImageContainer(img *StageImage) *StageImageContainer {
	c := &StageImageContainer{}
	c.image = img
	c.name = fmt.Sprintf("%s%v", image.StageContainerNamePrefix, util.GenerateConsistentRandomString(10))
	c.runOptions = newStageContainerOptions()
	c.commitChangeOptions = newStageContainerOptions()
	c.serviceCommitChangeOptions = newStageContainerOptions()
	return c
}

func (c *StageImageContainer) Name() string {
	return c.name
}

func (c *StageImageContainer) UserCommitChanges() []string {
	return c.commitChangeOptions.toCommitChanges()
}

func (c *StageImageContainer) UserRunCommands() []string {
	return c.runCommands
}

func (c *StageImageContainer) AddRunCommands(commands ...string) {
	c.runCommands = append(c.runCommands, commands...)
}

func (c *StageImageContainer) AddServiceRunCommands(commands ...string) {
	c.serviceRunCommands = append(c.serviceRunCommands, commands...)
}

func (c *StageImageContainer) RunOptions() ContainerOptions {
	return c.runOptions
}

func (c *StageImageContainer) CommitChangeOptions() ContainerOptions {
	return c.commitChangeOptions
}

func (c *StageImageContainer) ServiceCommitChangeOptions() ContainerOptions {
	return c.serviceCommitChangeOptions
}

func (c *StageImageContainer) prepareRunArgs() ([]string, error) {
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

	setColumnsEnv := fmt.Sprintf("--env=COLUMNS=%d", logboek.ContentWidth())
	runArgs = append(runArgs, setColumnsEnv)

	fromImageId := c.image.fromImage.GetID()

	args = append(args, runArgs...)
	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, c.prepareRunCommand())

	return args, nil
}

func (c *StageImageContainer) prepareRunCommand() string {
	return ShelloutPack(strings.Join(c.prepareRunCommands(), " && "))
}

func (c *StageImageContainer) prepareRunCommands() []string {
	runCommands := c.prepareAllRunCommands()
	if len(runCommands) != 0 {
		return runCommands
	} else {
		return []string{stapel.TrueBinPath()}
	}
}

func (c *StageImageContainer) prepareAllRunCommands() []string {
	var commands []string

	if debugDockerRunCommand() {
		commands = append(commands, "set -x")
	}

	commands = append(commands, c.serviceRunCommands...)
	commands = append(commands, c.runCommands...)

	return commands
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | %s --decode)", base64.StdEncoding.EncodeToString([]byte(command)), stapel.Base64BinPath())
}

func (c *StageImageContainer) prepareIntrospectBeforeArgs() ([]string, error) {
	args, err := c.prepareIntrospectArgsBase()
	if err != nil {
		return nil, err
	}

	fromImageId := c.image.fromImage.GetID()

	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, stapel.BashBinPath())

	return args, nil
}

func (c *StageImageContainer) prepareIntrospectArgs() ([]string, error) {
	args, err := c.prepareIntrospectArgsBase()
	if err != nil {
		return nil, err
	}

	imageId := c.image.GetID()

	args = append(args, imageId)
	args = append(args, "-ec")
	args = append(args, stapel.BashBinPath())

	return args, nil
}

func (c *StageImageContainer) prepareIntrospectArgsBase() ([]string, error) {
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

func (c *StageImageContainer) prepareRunOptions() (*StageImageContainerOptions, error) {
	serviceRunOptions, err := c.prepareServiceRunOptions()
	if err != nil {
		return nil, err
	}
	return serviceRunOptions.merge(c.runOptions), nil
}

func (c *StageImageContainer) prepareServiceRunOptions() (*StageImageContainerOptions, error) {
	serviceRunOptions := newStageContainerOptions()
	serviceRunOptions.Workdir = "/"
	serviceRunOptions.Entrypoint = stapel.BashBinPath()
	serviceRunOptions.User = "0:0"

	stapelContainerName, err := stapel.GetOrCreateContainer()
	if err != nil {
		return nil, err
	}

	serviceRunOptions.VolumesFrom = []string{stapelContainerName}

	return serviceRunOptions, nil
}

func (c *StageImageContainer) prepareIntrospectOptions() (*StageImageContainerOptions, error) {
	return c.prepareRunOptions()
}

func (c *StageImageContainer) prepareCommitChanges() ([]string, error) {
	commitOptions, err := c.prepareCommitOptions()
	if err != nil {
		return nil, err
	}

	commitChanges, err := commitOptions.prepareCommitChanges()
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *StageImageContainer) prepareCommitOptions() (*StageImageContainerOptions, error) {
	inheritedCommitOptions, err := c.prepareInheritedCommitOptions()
	if err != nil {
		return nil, err
	}

	commitOptions := inheritedCommitOptions.merge(c.serviceCommitChangeOptions.merge(c.commitChangeOptions))
	return commitOptions, nil
}

func (c *StageImageContainer) prepareInheritedCommitOptions() (*StageImageContainerOptions, error) {
	inheritedOptions := newStageContainerOptions()

	if c.image.fromImage == nil {
		panic(fmt.Sprintf("runtime error: FromImage should be (%s)", c.image.name))
	}

	if err := c.image.fromImage.MustResetInspect(); err != nil {
		return nil, fmt.Errorf("unable to reset inspect for image %s: %s", c.image.fromImage.Name(), err)
	}
	fromImageInspect := c.image.fromImage.GetInspect()

	if len(fromImageInspect.Config.Cmd) != 0 {
		inheritedOptions.Cmd = fmt.Sprintf("[\"%s\"]", strings.Join(fromImageInspect.Config.Cmd, "\", \""))
	}

	if len(fromImageInspect.Config.Entrypoint) != 0 {
		inheritedOptions.Entrypoint = fmt.Sprintf("[\"%s\"]", strings.Join(fromImageInspect.Config.Entrypoint, "\", \""))
	}

	inheritedOptions.User = fromImageInspect.Config.User
	if fromImageInspect.Config.WorkingDir != "" {
		inheritedOptions.Workdir = fromImageInspect.Config.WorkingDir
	} else {
		inheritedOptions.Workdir = "/"
	}
	return inheritedOptions, nil
}

func (c *StageImageContainer) run() error {
	runArgs, err := c.prepareRunArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun_LiveOutput(runArgs...); err != nil {
		return fmt.Errorf("container run failed: %s", err.Error())
	}

	return nil
}

func (c *StageImageContainer) introspect() error {
	runArgs, err := c.prepareIntrospectArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun_LiveOutput(runArgs...); err != nil {
		if !strings.Contains(err.Error(), "Code: ") || IsStartContainerErr(err) {
			return err
		}
	}

	return nil
}

func (c *StageImageContainer) introspectBefore() error {
	runArgs, err := c.prepareIntrospectBeforeArgs()
	if err != nil {
		return err
	}

	if err := docker.CliRun_LiveOutput(runArgs...); err != nil {
		if !strings.Contains(err.Error(), "Code: ") || IsStartContainerErr(err) {
			return err
		}
	}

	return nil
}

// https://docs.docker.com/engine/reference/run/#exit-status
func IsStartContainerErr(err error) bool {
	for _, code := range []string{"125", "126", "127"} {
		if strings.HasPrefix(err.Error(), fmt.Sprintf("Code: %s", code)) {
			return true
		}
	}

	return false
}

func (c *StageImageContainer) commit() (string, error) {
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

func (c *StageImageContainer) rm() error {
	return docker.ContainerRemove(c.name, types.ContainerRemoveOptions{})
}
