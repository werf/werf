package buildah

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah/types"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/werf"
)

type DockerWithFuseBuildah struct {
	BaseBuildah
}

func NewDockerWithFuseBuildah(commonOpts CommonBuildahOpts, opts DockerWithFuseModeOpts) (*DockerWithFuseBuildah, error) {
	b := &DockerWithFuseBuildah{}

	baseBuildah, err := NewBaseBuildah(commonOpts.TmpDir, BaseBuildahOpts{Insecure: commonOpts.Insecure})
	if err != nil {
		return nil, fmt.Errorf("unable to create BaseBuildah: %s", err)
	}
	b.BaseBuildah = *baseBuildah

	return b, nil
}

func (b *DockerWithFuseBuildah) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, []string{"tag", ref, newRef}, opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) Push(ctx context.Context, ref string, opts PushOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, []string{
		"push", fmt.Sprintf("--tls-verify=%s", strconv.FormatBool(!b.Insecure)), ref, fmt.Sprintf("docker://%s", ref),
	}, opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) Mount(ctx context.Context, container string, opts MountOpts) (string, error) {
	panic("not implemented")
}

func (b *DockerWithFuseBuildah) Umount(ctx context.Context, container string, opts UmountOpts) error {
	panic("not implemented")
}

func (b *DockerWithFuseBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	sessionTmpDir, _, _, err := b.prepareBuildFromDockerfile(dockerfile, opts.ContextTar)
	if err != nil {
		return "", fmt.Errorf("error preparing for build from dockerfile: %s", err)
	}
	defer func() {
		if debug() {
			return
		}

		if err = os.RemoveAll(sessionTmpDir); err != nil {
			logboek.Warn().LogF("unable to remove session tmp dir %q: %s\n", sessionTmpDir, err)
		}
	}()

	var buildArgs []string
	for k, v := range opts.BuildArgs {
		buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	// NOTE: it is principal to use cli option --tls-verify=true|false form with equality sign, instead of separate arguments (--tls-verify true|false), because --tls-verify is by itself a boolean argument
	budArgs := []string{"bud", "--format", "docker", fmt.Sprintf("--tls-verify=%s", strconv.FormatBool(!b.Insecure))}
	budArgs = append(budArgs, buildArgs...)
	budArgs = append(budArgs, "-f", "/.werf/buildah/tmp/Dockerfile")

	output, _, err := b.runBuildah(
		ctx,
		[]string{
			"--volume", fmt.Sprintf("%s:/.werf/buildah/tmp", sessionTmpDir),
			"--workdir", "/.werf/buildah/tmp/context",
		},
		budArgs, opts.LogWriter,
	)
	if err != nil {
		return "", err
	}

	outputLines := scanLines(output)

	return outputLines[len(outputLines)-1], nil
}

func (b *DockerWithFuseBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, append([]string{"run", container}, command...), opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) (string, error) {
	_, _, err := b.runBuildah(ctx, []string{}, []string{
		"from", fmt.Sprintf("--tls-verify=%s", strconv.FormatBool(!b.Insecure)), "--name", container, image,
	}, opts.LogWriter)
	// FIXME: return container name
	return "", err
}

// TODO(ilya-lesikov): make it more generic to handle not only images
func (b *DockerWithFuseBuildah) Inspect(ctx context.Context, ref string) (*types.BuilderInfo, error) {
	stdout, stderr, err := b.runBuildah(ctx, []string{}, []string{"inspect", "--type", "image", ref}, nil)
	if err != nil {
		if strings.Contains(stderr, "image not known") {
			return nil, nil
		}
		return nil, err
	}

	var res types.BuilderInfo
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		return nil, fmt.Errorf("unable to unmarshal buildah inspect json output: %s", err)
	}

	return &res, nil
}

func (b *DockerWithFuseBuildah) Pull(ctx context.Context, ref string, opts PullOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, []string{"pull", fmt.Sprintf("--tls-verify=%s", strconv.FormatBool(!b.Insecure)), ref}, opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	args := []string{"rmi"}
	if opts.Force {
		args = append(args, "-f")
	}
	args = append(args, ref)

	_, _, err := b.runBuildah(ctx, []string{}, args, opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) runBuildah(ctx context.Context, dockerArgs []string, buildahArgs []string, logWriter io.Writer) (string, string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	var stdoutWriter io.Writer
	var stderrWriter io.Writer

	if logWriter != nil {
		stdoutWriter = io.MultiWriter(stdout, logWriter)
		stderrWriter = io.MultiWriter(stderr, logWriter)
	} else {
		stdoutWriter = stdout
		stderrWriter = stderr
	}

	if err := runStorageContainer(ctx, BuildahStorageContainerName, BuildahImage); err != nil {
		return "", "", fmt.Errorf("unable to run werf buildah storage container: %s", err)
	}

	args := []string{"--rm"}
	args = append(args, dockerArgs...)
	args = append(args, BuildahWithFuseDockerArgs(BuildahStorageContainerName, docker.DockerConfigDir)...)
	args = append(args, buildahArgs...)

	if debug() {
		fmt.Printf("DEBUG CMD: docker run -ti %s\n", strings.Join(args, " "))
	}

	err := docker.CliRun_ProvidedOutput(ctx, stdoutWriter, stderrWriter, args...)
	return stdout.String(), stderr.String(), err
}

func runStorageContainer(ctx context.Context, name, image string) error {
	exist, err := docker.ContainerExist(ctx, name)
	if err != nil {
		return fmt.Errorf("unable to check existence of docker container %q: %s", name, err)
	}
	if exist {
		return nil
	}

	return werf.WithHostLock(ctx, fmt.Sprintf("buildah.container.%s", name), lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		return logboek.Context(ctx).LogProcess("Creating container %s using image %s", name, image).DoError(func() error {
			exist, err := docker.ContainerExist(ctx, name)
			if err != nil {
				return fmt.Errorf("unable to check existence of docker container %q: %s", name, err)
			}
			if exist {
				return nil
			}

			imageExist, err := docker.ImageExist(ctx, image)
			if err != nil {
				return fmt.Errorf("unable to check existence of docker image %q: %s", image, err)
			}
			if !imageExist {
				if err := docker.CliPullWithRetries(ctx, image); err != nil {
					return err
				}
			}

			return docker.CliCreate(ctx, "--name", name, image)
		})
	})
}

func BuildahWithFuseDockerArgs(storageContainerName, dockerConfigDir string) []string {
	return []string{
		"--user", "1000",
		"--device", "/dev/fuse",
		"--security-opt", "seccomp=unconfined",
		"--security-opt", "apparmor=unconfined",
		"--volume", fmt.Sprintf("%s:%s", dockerConfigDir, "/home/build/.docker"),
		"--volumes-from", storageContainerName,
		BuildahImage, "buildah",
	}
}

func scanLines(data string) []string {
	var lines []string

	s := bufio.NewScanner(strings.NewReader(data))
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines
}
