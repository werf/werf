package container_backend

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
)

type LegacyStageImageContainer struct {
	image                      *LegacyStageImage
	name                       string
	runCommands                []string
	serviceRunCommands         []string
	runOptions                 *LegacyStageImageContainerOptions
	commitChangeOptions        *LegacyStageImageContainerOptions
	serviceCommitChangeOptions *LegacyStageImageContainerOptions
}

func newLegacyStageImageContainer(img *LegacyStageImage) *LegacyStageImageContainer {
	c := &LegacyStageImageContainer{}
	c.image = img
	c.name = fmt.Sprintf("%s%v", image.StageContainerNamePrefix, util.GenerateConsistentRandomString(10))
	c.runOptions = newLegacyStageContainerOptions()
	c.commitChangeOptions = newLegacyStageContainerOptions()
	c.serviceCommitChangeOptions = newLegacyStageContainerOptions()
	return c
}

func (c *LegacyStageImageContainer) Name() string {
	return c.name
}

func (c *LegacyStageImageContainer) UserCommitChanges() []string {
	return c.commitChangeOptions.toCommitChanges(c.image.commitChangeOptions)
}

func (c *LegacyStageImageContainer) UserRunCommands() []string {
	return c.runCommands
}

func (c *LegacyStageImageContainer) AddRunCommands(commands ...string) {
	c.runCommands = append(c.runCommands, commands...)
}

func (c *LegacyStageImageContainer) AddServiceRunCommands(commands ...string) {
	c.serviceRunCommands = append(c.serviceRunCommands, commands...)
}

func (c *LegacyStageImageContainer) ServiceRunCommands(ctx context.Context) ([]string, error) {
	serviceRunCommands := c.serviceRunCommands
	{
		fromImageInspect, err := c.getFromImageInspect(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get from image inspect: %w", err)
		}

		for _, e := range fromImageInspect.Config.Env {
			pair := strings.SplitN(e, "=", 2)
			if pair[0] == "LD_LIBRARY_PATH" {
				serviceRunCommands = append(serviceRunCommands, fmt.Sprintf("export %q", e))
			}
		}
	}

	return serviceRunCommands, nil
}

func (c *LegacyStageImageContainer) RunOptions() LegacyContainerOptions {
	return c.runOptions
}

func (c *LegacyStageImageContainer) CommitChangeOptions() LegacyContainerOptions {
	return c.commitChangeOptions
}

func (c *LegacyStageImageContainer) ServiceCommitChangeOptions() LegacyContainerOptions {
	return c.serviceCommitChangeOptions
}

func (c *LegacyStageImageContainer) prepareRunArgs(ctx context.Context, runCommand string) ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.name))

	runOptions, err := c.prepareRunOptions(ctx)
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	setColumnsEnv := fmt.Sprintf("--env=COLUMNS=%d", logboek.Context(ctx).Streams().ContentWidth())
	runArgs = append(runArgs, setColumnsEnv)

	fromImageId := c.image.fromImage.GetID()

	args = append(args, runArgs...)
	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, runCommand)

	return args, nil
}

func (c *LegacyStageImageContainer) prepareAllRunCommands(ctx context.Context) ([]string, error) {
	var commands []string

	if debugDockerRunCommand() {
		commands = append(commands, "set -x")
	}

	serviceRunCommands, err := c.ServiceRunCommands(ctx)
	if err != nil {
		return nil, err
	}

	commands = append(commands, serviceRunCommands...)
	commands = append(commands, c.runCommands...)

	if len(commands) == 0 {
		return []string{stapel.TrueBinPath()}, nil
	}

	return commands, nil
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | %s --decode)", base64.StdEncoding.EncodeToString([]byte(command)), stapel.Base64BinPath())
}

func (c *LegacyStageImageContainer) prepareIntrospectBeforeArgs(ctx context.Context) ([]string, error) {
	args, err := c.prepareIntrospectArgsBase(ctx)
	if err != nil {
		return nil, err
	}

	fromImageId := c.image.fromImage.GetID()

	args = append(args, fromImageId)
	args = append(args, "-ec")
	args = append(args, stapel.BashBinPath())

	return args, nil
}

func (c *LegacyStageImageContainer) prepareIntrospectArgs(ctx context.Context) ([]string, error) {
	args, err := c.prepareIntrospectArgsBase(ctx)
	if err != nil {
		return nil, err
	}

	imageId := c.image.GetID()

	args = append(args, imageId)
	args = append(args, "-ec")
	args = append(args, stapel.BashBinPath())

	return args, nil
}

func (c *LegacyStageImageContainer) prepareIntrospectArgsBase(ctx context.Context) ([]string, error) {
	var args []string

	runOptions, err := c.prepareIntrospectOptions(ctx)
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

func (c *LegacyStageImageContainer) prepareRunOptions(ctx context.Context) (*LegacyStageImageContainerOptions, error) {
	serviceRunOptions, err := c.prepareServiceRunOptions(ctx)
	if err != nil {
		return nil, err
	}
	return serviceRunOptions.merge(c.runOptions), nil
}

func (c *LegacyStageImageContainer) prepareServiceRunOptions(ctx context.Context) (*LegacyStageImageContainerOptions, error) {
	serviceRunOptions := newLegacyStageContainerOptions()
	serviceRunOptions.Workdir = "/"
	serviceRunOptions.Entrypoint = stapel.BashBinPath()
	serviceRunOptions.User = "0:0"
	serviceRunOptions.Env["LD_LIBRARY_PATH"] = ""

	stapelContainerName, err := stapel.GetOrCreateContainer(ctx)
	if err != nil {
		return nil, err
	}

	serviceRunOptions.VolumesFrom = []string{stapelContainerName}

	return serviceRunOptions, nil
}

func (c *LegacyStageImageContainer) prepareIntrospectOptions(ctx context.Context) (*LegacyStageImageContainerOptions, error) {
	return c.prepareRunOptions(ctx)
}

func (c *LegacyStageImageContainer) prepareCommitChanges(ctx context.Context, opts LegacyCommitChangeOptions) ([]string, error) {
	commitOptions, err := c.prepareCommitOptions(ctx)
	if err != nil {
		return nil, err
	}

	commitChanges, err := commitOptions.prepareCommitChanges(ctx, opts)
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *LegacyStageImageContainer) prepareCommitOptions(ctx context.Context) (*LegacyStageImageContainerOptions, error) {
	inheritedCommitOptions, err := c.prepareInheritedCommitOptions(ctx)
	if err != nil {
		return nil, err
	}

	commitOptions := inheritedCommitOptions.merge(c.serviceCommitChangeOptions.merge(c.commitChangeOptions))
	return commitOptions, nil
}

func (c *LegacyStageImageContainer) prepareInheritedCommitOptions(ctx context.Context) (*LegacyStageImageContainerOptions, error) {
	inheritedOptions := newLegacyStageContainerOptions()

	if c.image.fromImage == nil {
		panic(fmt.Sprintf("runtime error: FromImage should be (%s)", c.image.name))
	}

	if err := c.image.fromImage.MustResetInfo(ctx); err != nil {
		return nil, fmt.Errorf("unable to reset info for image %s: %w", c.image.fromImage.Name(), err)
	}

	fromImageInspect, err := c.getFromImageInspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get from image inspect: %w", err)
	}

	if len(fromImageInspect.Config.Cmd) != 0 {
		inheritedOptions.Cmd = fmt.Sprintf("[\"%s\"]", strings.Join(fromImageInspect.Config.Cmd, "\", \""))
	}

	if len(fromImageInspect.Config.Entrypoint) != 0 {
		inheritedOptions.Entrypoint = fmt.Sprintf("[\"%s\"]", strings.Join(fromImageInspect.Config.Entrypoint, "\", \""))
	}

	for _, e := range fromImageInspect.Config.Env {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == "LD_LIBRARY_PATH" {
			inheritedOptions.Env[pair[0]] = pair[1]
		}
	}

	inheritedOptions.User = fromImageInspect.Config.User
	if fromImageInspect.Config.WorkingDir != "" {
		inheritedOptions.Workdir = fromImageInspect.Config.WorkingDir
	} else {
		inheritedOptions.Workdir = "/"
	}
	return inheritedOptions, nil
}

func (c *LegacyStageImageContainer) getFromImageInspect(ctx context.Context) (*types.ImageInspect, error) {
	dockerServerBackend := c.image.ContainerBackend.(*DockerServerBackend)
	fromImageInspect, err := dockerServerBackend.GetImageInspect(ctx, c.image.fromImage.Name())
	if err != nil {
		return nil, err
	}

	return fromImageInspect, nil
}

func (c *LegacyStageImageContainer) run(ctx context.Context, runArgs []string) error {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	RegisterRunningContainer(c.name, ctx)
	err := docker.CliRun_LiveOutput(ctx, runArgs...)
	UnregisterRunningContainer(c.name)
	if err != nil {
		return fmt.Errorf("container run failed: %w", err)
	}
	return nil
}

func (c *LegacyStageImageContainer) introspect(ctx context.Context) error {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	runArgs, err := c.prepareIntrospectArgs(ctx)
	if err != nil {
		return err
	}

	if err := docker.CliRun_LiveOutput(ctx, runArgs...); err != nil {
		if !strings.Contains(err.Error(), "Code: ") || IsStartContainerErr(err) {
			return err
		}
	}

	return nil
}

func (c *LegacyStageImageContainer) introspectBefore(ctx context.Context) error {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	runArgs, err := c.prepareIntrospectBeforeArgs(ctx)
	if err != nil {
		return err
	}

	if err := docker.CliRun_LiveOutput(ctx, runArgs...); err != nil {
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

func (c *LegacyStageImageContainer) commit(ctx context.Context) (string, error) {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	commitChanges, err := c.prepareCommitChanges(ctx, c.image.commitChangeOptions)
	if err != nil {
		return "", err
	}

	commitOptions := types.ContainerCommitOptions{Changes: commitChanges}
	id, err := docker.ContainerCommit(ctx, c.name, commitOptions)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *LegacyStageImageContainer) rm(ctx context.Context) error {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	err := docker.ContainerRemove(ctx, c.name, types.ContainerRemoveOptions{RemoveVolumes: true, Force: true})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("removal of container %s is already in progress", c.name)) {
			return nil
		}
		return fmt.Errorf("unable to remove container %s: %w", c.name, err)
	}
	return nil
}
