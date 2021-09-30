package buildah

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/werf/pkg/buildah/types"
	"github.com/werf/werf/pkg/werf"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
)

type DockerWithFuseBuildah struct {
	BaseBuildah
}

func NewDockerWithFuseBuildah() (*DockerWithFuseBuildah, error) {
	b := &DockerWithFuseBuildah{}

	baseBuildah, err := NewBaseBuildah()
	if err != nil {
		return nil, fmt.Errorf("unable to create BaseBuildah: %s", err)
	}
	b.BaseBuildah = *baseBuildah

	return b, nil
}

func (b *DockerWithFuseBuildah) Tag(ctx context.Context, ref, newRef string) error {
	panic("not implemented")
}
func (b *DockerWithFuseBuildah) Push(ctx context.Context, ref string, opts PushOpts) error {
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

	output, _, err := b.runBuildah(
		ctx,
		[]string{
			"--volume", fmt.Sprintf("%s:/.werf/buildah/tmp", sessionTmpDir),
			"--workdir", "/.werf/buildah/tmp/context",
		},
		[]string{"bud", "-f", "/.werf/buildah/tmp/Dockerfile"}, opts.LogWriter,
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

func (b *DockerWithFuseBuildah) FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, []string{"from", "--name", container, image}, opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) Inspect(ctx context.Context, ref string) (types.BuilderInfo, error) {
	stdout, _, err := b.runBuildah(ctx, []string{}, []string{"inspect", ref}, nil)
	if err != nil {
		return types.BuilderInfo{}, nil
	}

	var res types.BuilderInfo

	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		return types.BuilderInfo{}, fmt.Errorf("unable to unmarshal buildah inspect json output: %s", err)
	}

	return res, nil
}

func (b *DockerWithFuseBuildah) Pull(ctx context.Context, ref string, opts PullOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, []string{"pull", ref}, opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) Rmi(ctx context.Context, ref string) error {
	panic("not implemented yet")
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
	args = append(args, buildahWithFuseDockerArgs(BuildahStorageContainerName)...)
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
		return fmt.Errorf("unable to check existance of docker container %q: %s", name, err)
	}
	if exist {
		return nil
	}

	return werf.WithHostLock(ctx, fmt.Sprintf("buildah.container.%s", name), lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		return logboek.Context(ctx).LogProcess("Creating container %s using image %s", name, image).DoError(func() error {
			exist, err := docker.ContainerExist(ctx, name)
			if err != nil {
				return fmt.Errorf("unable to check existance of docker container %q: %s", name, err)
			}
			if exist {
				return nil
			}

			imageExist, err := docker.ImageExist(ctx, image)
			if err != nil {
				return fmt.Errorf("unable to check existance of docker image %q: %s", image, err)
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

func buildahWithFuseDockerArgs(storageContainerName string) []string {
	return []string{
		"--user", "1000",
		"--device", "/dev/fuse",
		"--security-opt", "seccomp=unconfined",
		"--security-opt", "apparmor=unconfined",
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
