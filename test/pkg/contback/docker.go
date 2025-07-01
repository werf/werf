package contback

import (
	"bytes"
	"context"
	"encoding/json"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

func NewDockerBackend() ContainerBackend {
	return &DockerBackend{}
}

type DockerBackend struct {
	BaseContainerBackend
}

func (r *DockerBackend) ExpectCmdsToSucceed(ctx context.Context, image string, cmds ...string) {
	expectCmdsToSucceed(ctx, r, image, cmds...)
}

func (r *DockerBackend) RunSleepingContainer(ctx context.Context, containerName, image string) {
	args := r.CommonCliArgs
	args = append(args, "run", "--rm", "-d", "--entrypoint=", "--name", containerName, image, "sleep", "infinity")
	utils.RunSucceedCommand(ctx, "/", "docker", args...)
}

func (r *DockerBackend) Exec(ctx context.Context, containerName string, cmds ...string) {
	for _, cmd := range cmds {
		args := r.CommonCliArgs
		args = append(args, "exec", containerName, "bash", "-o", "pipefail", "-euc", cmd)
		utils.RunSucceedCommand(ctx, "/", "docker", args...)
	}
}

func (r *DockerBackend) Rm(ctx context.Context, containerName string) {
	args := r.CommonCliArgs
	args = append(args, "rm", "-fv", containerName)
	utils.RunSucceedCommand(ctx, "/", "docker", args...)
}

func (r *DockerBackend) Pull(ctx context.Context, image string) {
	args := r.CommonCliArgs
	args = append(args, "pull", image)
	utils.RunSucceedCommand(ctx, "/", "docker", args...)
}

func (r *DockerBackend) DumpImage(ctx context.Context, image string) *bytes.Reader {
	args := r.CommonCliArgs
	args = append(args, "image", "save", image)

	b, err := utils.RunCommandWithOptions(ctx, "/", "docker", args, utils.RunCommandOptions{
		NoStderr:      true,
		ShouldSucceed: true,
	})

	Expect(err).NotTo(HaveOccurred())

	return bytes.NewReader(b)
}

func (r *DockerBackend) GetImageInspect(ctx context.Context, image string) DockerImageInspect {
	args := r.CommonCliArgs
	args = append(args, "image", "inspect", image)
	inspectRaw, err := utils.RunCommand(ctx, "/", "docker", args...)

	var dockerInspect DockerInspect

	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(inspectRaw, &dockerInspect)).To(Succeed())
	Expect(len(dockerInspect)).To(Equal(1))

	return dockerInspect[0]
}

type DockerInspect []DockerImageInspect
