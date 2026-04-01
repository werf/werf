package docker

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	dockerCli "github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	containerCmd "github.com/docker/cli/cli/command/container"
	"github.com/docker/docker/api/types"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/net/context"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend/thirdparty/platformutil"
)

func Containers(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return apiCli(ctx).ContainerList(ctx, options)
}

func ContainerExist(ctx context.Context, ref string) (bool, error) {
	if _, err := ContainerInspect(ctx, ref); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ContainerAttach(ctx context.Context, ref string, options types.ContainerAttachOptions) (types.HijackedResponse, error) {
	return apiCli(ctx).ContainerAttach(ctx, ref, options)
}

func ContainerInspect(ctx context.Context, ref string) (types.ContainerJSON, error) {
	return apiCli(ctx).ContainerInspect(ctx, ref)
}

func ContainerCommit(ctx context.Context, ref string, commitOptions types.ContainerCommitOptions) (string, error) {
	response, err := apiCli(ctx).ContainerCommit(ctx, ref, commitOptions)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func ContainerRemove(ctx context.Context, ref string, options types.ContainerRemoveOptions) error {
	return apiCli(ctx).ContainerRemove(ctx, ref, options)
}

func CliCreate(ctx context.Context, args ...string) error {
	var containerName string
	var imageName string
	var rawVolumes []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "--name=") {
			containerName = strings.TrimPrefix(arg, "--name=")
		} else if strings.HasPrefix(arg, "--volume=") {
			rawVolumes = append(rawVolumes, strings.TrimPrefix(arg, "--volume="))
		} else if !strings.HasPrefix(arg, "-") {
			imageName = arg
		}
	}

	if imageName == "" {
		return fmt.Errorf("create requires image name")
	}

	var binds []string
	configVolumes := map[string]struct{}{}
	for _, v := range rawVolumes {
		if strings.Contains(v, ":") {
			binds = append(binds, v)
		} else {
			configVolumes[v] = struct{}{}
		}
	}

	config := &containerType.Config{Image: imageName}
	if len(configVolumes) > 0 {
		config.Volumes = configVolumes
	}

	_, err := apiCli(ctx).ContainerCreate(ctx,
		config,
		&containerType.HostConfig{Binds: binds},
		nil,
		nil,
		containerName,
	)
	return err
}

func doCliRun(ctx context.Context, c command.Cli, args ...string) error {
	return prepareCliCmd(ctx, containerCmd.NewRunCommand(c), args...).Execute()
}

type cliRunOutputMode int

const (
	cliRunOutputAuto cliRunOutputMode = iota
	cliRunOutputLive
	cliRunOutputRecorded
)

type cliRunOutputModeKey struct{}

func withCliRunOutputMode(ctx context.Context, mode cliRunOutputMode) context.Context {
	return context.WithValue(ctx, cliRunOutputModeKey{}, mode)
}

func cliRunOutputModeFromContext(ctx context.Context) cliRunOutputMode {
	mode, ok := ctx.Value(cliRunOutputModeKey{}).(cliRunOutputMode)
	if !ok {
		return cliRunOutputAuto
	}
	return mode
}

func ensureLogboekContext(ctx context.Context) context.Context {
	if ctx.Value("logboek_logger") != nil {
		return ctx
	}
	return logboek.NewContext(ctx, logboek.DefaultLogger())
}

func doCliRunSDK(ctx context.Context, args ...string) (string, int, error) {
	ctx = ensureLogboekContext(ctx)

	var (
		autoRemove    bool
		containerName string
		binds         []string
		entrypoint    string
		env           []string
		user          string
		workdir       string
		platformStr   string
		volumesFrom   []string
		imageName     string
		cmdArgs       []string
		detach        bool
		exposedPorts  []string
		networkMode   string
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if imageName != "" {
			cmdArgs = append(cmdArgs, args[i:]...)
			break
		}

		switch {
		case arg == "--rm":
			autoRemove = true
		case strings.HasPrefix(arg, "--name="):
			containerName = strings.TrimPrefix(arg, "--name=")
		case arg == "--name":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --name requires value")
			}
			i++
			containerName = args[i]
		case strings.HasPrefix(arg, "--volume="):
			binds = append(binds, strings.TrimPrefix(arg, "--volume="))
		case arg == "--volume":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --volume requires value")
			}
			i++
			binds = append(binds, args[i])
		case strings.HasPrefix(arg, "--entrypoint="):
			entrypoint = strings.TrimPrefix(arg, "--entrypoint=")
		case arg == "--entrypoint":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --entrypoint requires value")
			}
			i++
			entrypoint = args[i]
		case strings.HasPrefix(arg, "--env="):
			env = append(env, strings.TrimPrefix(arg, "--env="))
		case arg == "--env":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --env requires value")
			}
			i++
			env = append(env, args[i])
		case strings.HasPrefix(arg, "--user="):
			user = strings.TrimPrefix(arg, "--user=")
		case arg == "--user":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --user requires value")
			}
			i++
			user = args[i]
		case strings.HasPrefix(arg, "--workdir="):
			workdir = strings.TrimPrefix(arg, "--workdir=")
		case arg == "--workdir":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --workdir requires value")
			}
			i++
			workdir = args[i]
		case strings.HasPrefix(arg, "--platform="):
			platformStr = strings.TrimPrefix(arg, "--platform=")
		case arg == "--platform":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --platform requires value")
			}
			i++
			platformStr = args[i]
		case strings.HasPrefix(arg, "--volumes-from="):
			volumesFrom = append(volumesFrom, strings.TrimPrefix(arg, "--volumes-from="))
		case arg == "--volumes-from":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --volumes-from requires value")
			}
			i++
			volumesFrom = append(volumesFrom, args[i])
		case arg == "--detach" || arg == "-d":
			detach = true
		case strings.HasPrefix(arg, "--expose="):
			exposedPorts = append(exposedPorts, strings.TrimPrefix(arg, "--expose="))
		case arg == "--expose":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --expose requires value")
			}
			i++
			exposedPorts = append(exposedPorts, args[i])
		case strings.HasPrefix(arg, "--network="):
			networkMode = strings.TrimPrefix(arg, "--network=")
		case arg == "--network":
			if i+1 >= len(args) {
				return "", -1, fmt.Errorf("flag --network requires value")
			}
			i++
			networkMode = args[i]
		case strings.HasPrefix(arg, "-"):
			return "", -1, fmt.Errorf("unsupported docker run flag %q", arg)
		default:
			imageName = arg
		}
	}

	if imageName == "" {
		return "", -1, fmt.Errorf("run requires image name")
	}

	config := &containerType.Config{
		Image: imageName,
		Cmd:   cmdArgs,
	}
	if entrypoint != "" {
		config.Entrypoint = []string{entrypoint}
	}
	if len(env) > 0 {
		config.Env = env
	}
	if user != "" {
		config.User = user
	}
	if workdir != "" {
		config.WorkingDir = workdir
	}
	if len(exposedPorts) > 0 {
		config.ExposedPorts = nat.PortSet{}
		for _, p := range exposedPorts {
			port, err := nat.NewPort("tcp", p)
			if err != nil {
				return "", -1, fmt.Errorf("parse expose port %q: %w", p, err)
			}
			config.ExposedPorts[port] = struct{}{}
		}
	}

	removeAfterRun := autoRemove && !detach

	hostConfig := &containerType.HostConfig{
		Binds:       binds,
		VolumesFrom: volumesFrom,
	}
	if detach && autoRemove {
		hostConfig.AutoRemove = true
	}
	if networkMode != "" {
		hostConfig.NetworkMode = containerType.NetworkMode(networkMode)
	}

	var platform *specs.Platform
	if platformStr != "" {
		parsed, err := platformutil.ParsePlatform(platformStr)
		if err != nil {
			return "", -1, fmt.Errorf("parse platform %q: %w", platformStr, err)
		}
		platform = &parsed
	}

	createResp, err := apiCli(ctx).ContainerCreate(ctx, config, hostConfig, nil, platform, containerName)
	if err != nil {
		return "", -1, err
	}
	if removeAfterRun {
		defer func() {
			_ = ContainerRemove(ctx, createResp.ID, types.ContainerRemoveOptions{Force: true})
		}()
	}

	if err := apiCli(ctx).ContainerStart(ctx, createResp.ID, containerType.StartOptions{}); err != nil {
		return "", -1, err
	}

	if detach {
		return createResp.ID, 0, nil
	}

	statusCh, errCh := apiCli(ctx).ContainerWait(ctx, createResp.ID, containerType.WaitConditionNotRunning)

	var output bytes.Buffer
	stdoutWriter := io.Writer(&output)
	stderrWriter := io.Writer(&output)

	switch cliRunOutputModeFromContext(ctx) {
	case cliRunOutputLive:
		stdoutWriter = io.MultiWriter(stdoutWriter, logboek.Context(ctx).OutStream())
		stderrWriter = io.MultiWriter(stderrWriter, logboek.Context(ctx).ErrStream())
	case cliRunOutputRecorded:
	default:
		if liveCliOutputEnabled {
			stdoutWriter = io.MultiWriter(stdoutWriter, logboek.Context(ctx).OutStream())
			stderrWriter = io.MultiWriter(stderrWriter, logboek.Context(ctx).ErrStream())
		}
	}

	var logErr error
	if logReader, err := apiCli(ctx).ContainerLogs(ctx, createResp.ID, containerType.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true}); err != nil {
		logErr = err
	} else {
		defer logReader.Close()
		_, logErr = stdcopy.StdCopy(stdoutWriter, stderrWriter, logReader)
	}

	exitCode := -1
	for exitCode == -1 {
		select {
		case waitErr := <-errCh:
			if waitErr != nil {
				return output.String(), exitCode, waitErr
			}
		case status := <-statusCh:
			exitCode = int(status.StatusCode)
		}
	}

	if exitCode != 0 {
		return output.String(), exitCode, dockerCli.StatusError{StatusCode: exitCode, Status: fmt.Sprintf("exit code %d", exitCode)}
	}
	if logErr != nil {
		return output.String(), exitCode, logErr
	}

	return output.String(), exitCode, nil
}

func CliRun(ctx context.Context, args ...string) error {
	ctx = ensureLogboekContext(ctx)
	ctx = withCliRunOutputMode(ctx, cliRunOutputAuto)
	output, _, err := doCliRunSDK(ctx, args...)
	if err != nil && !liveCliOutputEnabled {
		logboek.Context(ctx).Warn().LogF("%s", output)
	}
	return err
}

func CliRun_ProvidedOutput(ctx context.Context, stdoutWriter, stderrWriter io.Writer, args ...string) error {
	return callCliWithProvidedOutput(ctx, stdoutWriter, stderrWriter, func(c command.Cli) error {
		return doCliRun(ctx, c, args...)
	})
}

func CliRun_LiveOutput(ctx context.Context, args ...string) error {
	ctx = ensureLogboekContext(ctx)
	ctx = withCliRunOutputMode(ctx, cliRunOutputLive)
	_, _, err := doCliRunSDK(ctx, args...)
	return err
}

func CliRun_RecordedOutput(ctx context.Context, args ...string) (string, error) {
	ctx = ensureLogboekContext(ctx)
	ctx = withCliRunOutputMode(ctx, cliRunOutputRecorded)
	output, _, err := doCliRunSDK(ctx, args...)
	return output, err
}

func CliRm(ctx context.Context, args ...string) error {
	force := false
	containerRefs := []string{}

	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else {
			containerRefs = append(containerRefs, arg)
		}
	}

	for _, ref := range containerRefs {
		if err := ContainerRemove(ctx, ref, types.ContainerRemoveOptions{Force: force}); err != nil {
			return err
		}
	}

	return nil
}

func CliRm_RecordedOutput(ctx context.Context, args ...string) (string, error) {
	if err := CliRm(ctx, args...); err != nil {
		return "", err
	}
	return "", nil
}
