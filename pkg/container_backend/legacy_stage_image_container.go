package container_backend

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/docker/cli/cli"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/stapel"
)

type LegacyStageImageContainer struct {
	image                      *LegacyStageImage
	name                       string
	runCommands                []string
	serviceRunCommands         []string
	runOptions                 *LegacyStageImageContainerOptions
	commitChangeOptions        *LegacyStageImageContainerOptions
	serviceCommitChangeOptions *LegacyStageImageContainerOptions
	inheritedCommitOptions     *LegacyStageImageContainerOptions
	buildTimeEnv               map[string]string
}

func newLegacyStageImageContainer(img *LegacyStageImage) *LegacyStageImageContainer {
	c := &LegacyStageImageContainer{}
	c.image = img
	c.name = fmt.Sprintf("%s%v", image.StageContainerNamePrefix, util.GenerateConsistentRandomString(10))
	c.runOptions = newLegacyStageContainerOptions()
	c.commitChangeOptions = newLegacyStageContainerOptions()
	c.serviceCommitChangeOptions = newLegacyStageContainerOptions()
	c.buildTimeEnv = make(map[string]string)
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

func (c *LegacyStageImageContainer) RunOptions() LegacyContainerOptions {
	return c.runOptions
}

func (c *LegacyStageImageContainer) CommitChangeOptions() LegacyContainerOptions {
	return c.commitChangeOptions
}

func (c *LegacyStageImageContainer) ServiceCommitChangeOptions() LegacyContainerOptions {
	return c.serviceCommitChangeOptions
}

func (c *LegacyStageImageContainer) AddBuildTimeEnv(envs map[string]string) {
	for k, v := range envs {
		c.buildTimeEnv[k] = v
	}
}

func (c *LegacyStageImageContainer) prepareRunArgs(ctx context.Context) ([]string, error) {
	var args []string
	args = append(args, fmt.Sprintf("--name=%s", c.name))

	if c.image.GetTargetPlatform() != "" {
		args = append(args, fmt.Sprintf("--platform=%s", c.image.GetTargetPlatform()))
	}

	runOptions, err := c.prepareRunOptions(ctx)
	if err != nil {
		return nil, err
	}

	runArgs, err := runOptions.toRunArgs()
	if err != nil {
		return nil, err
	}

	args = append(args, runArgs...)
	args = append(args, c.prepareRunCommandArgs()...)

	return args, nil
}

func shellSingleQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", `'\''`) + "'"
}

func (c *LegacyStageImageContainer) prepareBuildTimeEnvExports(ctx context.Context) []string {
	envs := make(map[string]string, len(c.buildTimeEnv)+1)
	for k, v := range c.buildTimeEnv {
		envs[k] = v
	}
	envs["COLUMNS"] = fmt.Sprintf("%d", logboek.Context(ctx).Streams().ContentWidth())

	var exports []string
	for _, k := range sortStrings(getKeys(envs)) {
		exports = append(exports, fmt.Sprintf("export %s=%s", k, shellSingleQuote(envs[k])))
	}
	return exports
}

func (c *LegacyStageImageContainer) prepareRunCommandArgs() []string {
	// The assignment reads the script to EOF with Bash alone, so it needs no binary from the
	// base or the stapel image, a reader failure aborts before eval, and build commands find
	// stdin already drained. The redirect only makes that explicit.
	return []string{
		"-i", c.imageRef(c.image.fromImage), "-ec",
		`script=$(</dev/stdin); eval "$script" < /dev/null`,
	}
}

func (c *LegacyStageImageContainer) prepareDebugRunCommand(ctx context.Context, runArgs []string) string {
	return fmt.Sprintf("printf '%%s' %s | docker run %s", shellescape.Quote(c.prepareRunCommand(ctx)), shellescape.QuoteCommand(runArgs))
}

func (c *LegacyStageImageContainer) prepareRunCommand(ctx context.Context) string {
	commands := append(c.prepareBuildTimeEnvExports(ctx), c.prepareRunCommands()...)
	return strings.Join(commands, " && ")
}

func (c *LegacyStageImageContainer) prepareRunCommands() []string {
	runCommands := c.prepareAllRunCommands()
	if len(runCommands) != 0 {
		return runCommands
	} else {
		return []string{stapel.TrueBinPath()}
	}
}

func (c *LegacyStageImageContainer) prepareAllRunCommands() []string {
	var commands []string

	if debugDockerRunCommand() {
		commands = append(commands, "set -x")
	}

	commands = append(commands, c.serviceRunCommands...)
	commands = append(commands, c.runCommands...)

	return commands
}

func (c *LegacyStageImageContainer) imageRef(img *LegacyStageImage) string {
	if img.BuiltID() != "" {
		return img.GetID()
	}

	if c.image.GetTargetPlatform() != "" {
		return img.Name()
	}

	return img.GetID()
}

func (c *LegacyStageImageContainer) prepareIntrospectBeforeArgs(ctx context.Context) ([]string, error) {
	args, err := c.prepareIntrospectArgsBase(ctx)
	if err != nil {
		return nil, err
	}

	args = append(args, c.imageRef(c.image.fromImage))
	args = append(args, "-ec")
	args = append(args, stapel.BashBinPath())

	return args, nil
}

func (c *LegacyStageImageContainer) prepareIntrospectArgs(ctx context.Context) ([]string, error) {
	args, err := c.prepareIntrospectArgsBase(ctx)
	if err != nil {
		return nil, err
	}

	args = append(args, c.imageRef(c.image))
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

	// ENV
	serviceRunOptions.Env["LANG"] = "C.UTF-8"
	serviceRunOptions.Env["LC_ALL"] = "C.UTF-8"

	stapelContainerName, err := stapel.GetOrCreateContainer(ctx, c.image.GetTargetPlatform())
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
	commitOptions, err := c.prepareCommitOptions()
	if err != nil {
		return nil, err
	}

	commitChanges, err := commitOptions.prepareCommitChanges(ctx, opts)
	if err != nil {
		return nil, err
	}
	return commitChanges, nil
}

func (c *LegacyStageImageContainer) prepareCommitOptions() (*LegacyStageImageContainerOptions, error) {
	if c.inheritedCommitOptions == nil {
		return nil, fmt.Errorf("commit options inherited from the base image are not prepared for image %s", c.image.name)
	}

	commitOptions := c.inheritedCommitOptions.merge(c.serviceCommitChangeOptions.merge(c.commitChangeOptions))
	return commitOptions, nil
}

// prepareInheritedCommitOptions reads the config of the base image to restore in the committed image
// what running the build container overrides. It must be called while the container has not been
// committed yet: the base image is guaranteed to be available locally only until then.
func (c *LegacyStageImageContainer) prepareInheritedCommitOptions(ctx context.Context, fromImageRef string) (*LegacyStageImageContainerOptions, error) {
	inheritedOptions := newLegacyStageContainerOptions()

	dockerServerBackend := c.image.ContainerBackend.(*DockerServerBackend)

	fromImageInspect, err := dockerServerBackend.GetImageInspect(ctx, fromImageRef)
	if err != nil {
		return nil, fmt.Errorf("unable to get image inspect: %w", err)
	}
	if fromImageInspect == nil {
		return nil, fmt.Errorf("image %s is not available locally", fromImageRef)
	}

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

	fromImageEnv := convertKVStringsToMap(fromImageInspect.Config.Env)
	for _, k := range []string{"LANG", "LC_ALL"} {
		if val, hasKey := fromImageEnv[k]; hasKey {
			inheritedOptions.Env[k] = val
		}
	}

	return inheritedOptions, nil
}

func (c *LegacyStageImageContainer) run(ctx context.Context) error {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	if c.image.fromImage == nil {
		panic(fmt.Sprintf("runtime error: FromImage should be (%s)", c.image.name))
	}

	inheritedCommitOptions, err := c.prepareInheritedCommitOptions(ctx, c.imageRef(c.image.fromImage))
	if err != nil {
		return err
	}
	c.inheritedCommitOptions = inheritedCommitOptions

	runArgs, err := c.prepareRunArgs(ctx)
	if err != nil {
		return err
	}

	RegisterRunningContainer(c.name, ctx)
	err = docker.CliRunWithInput_LiveOutput(ctx, c.prepareRunCommand(ctx), runArgs...)
	UnregisterRunningContainer(c.name)
	if err != nil {
		return fmt.Errorf("container run failed: %w", namedContainerExitErr(err))
	}
	return nil
}

func containerExitCode(err error) (int, bool) {
	var statusErr cli.StatusError
	if !errors.As(err, &statusErr) {
		return 0, false
	}

	return statusErr.StatusCode, true
}

// docker/cli reports a non-zero container exit as a cli.StatusError with an empty message, so the
// code has to be spelled out for the user. The empty verb keeps the error chain without a separator.
func namedContainerExitErr(err error) error {
	code, ok := containerExitCode(err)
	if !ok || err.Error() != "" {
		return err
	}

	return fmt.Errorf("exit code %d%w", code, err)
}

func (c *LegacyStageImageContainer) introspect(ctx context.Context) error {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	runArgs, err := c.prepareIntrospectArgs(ctx)
	if err != nil {
		return err
	}

	if err := docker.CliRun_LiveOutput(ctx, runArgs...); err != nil {
		if _, ok := containerExitCode(err); !ok || IsStartContainerErr(err) {
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
		if _, ok := containerExitCode(err); !ok || IsStartContainerErr(err) {
			return err
		}
	}

	return nil
}

// https://docs.docker.com/engine/reference/run/#exit-status
func IsStartContainerErr(err error) bool {
	code, ok := containerExitCode(err)

	return ok && lo.Contains([]int{125, 126, 127}, code)
}

func (c *LegacyStageImageContainer) commit(ctx context.Context) (string, error) {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	commitChanges, err := c.prepareCommitChanges(ctx, c.image.commitChangeOptions)
	if err != nil {
		return "", err
	}

	commitOptions := dockercontainer.CommitOptions{Changes: commitChanges}
	id, err := docker.ContainerCommit(ctx, c.name, commitOptions)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *LegacyStageImageContainer) rm(ctx context.Context) error {
	_ = c.image.ContainerBackend.(*DockerServerBackend)

	err := docker.ContainerRemove(ctx, c.name, dockercontainer.RemoveOptions{RemoveVolumes: true, Force: true})
	if err != nil {
		if errdefs.IsNotFound(err) || errdefs.IsConflict(err) {
			return nil
		}

		return fmt.Errorf("unable to remove container %s: %w", c.name, err)
	}
	return nil
}

func convertKVStringsToMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		k, v, _ := strings.Cut(value, "=")
		result[k] = v
	}
	return result
}
