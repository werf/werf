package contback

import (
	"encoding/json"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

func NewDockerBackend() ContainerBackend {
	return &DockerBackend{}
}

type DockerBackend struct {
	BaseContainerBackend
}

func (r *DockerBackend) ExpectCmdsToSucceed(image string, cmds ...string) {
	expectCmdsToSucceed(r, image, cmds...)
}

func (r *DockerBackend) RunSleepingContainer(containerName, image string) {
	args := r.CommonCliArgs
	args = append(args, "run", "--rm", "-d", "--entrypoint=", "--name", containerName, image, "tail", "-f", "/dev/null")
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerBackend) Exec(containerName string, cmds ...string) {
	for _, cmd := range cmds {
		args := r.CommonCliArgs
		args = append(args, "exec", containerName, "bash", "-o", "pipefail", "-euc", cmd)
		utils.RunSucceedCommand("/", "docker", args...)
	}
}

func (r *DockerBackend) Rm(containerName string) {
	args := r.CommonCliArgs
	args = append(args, "rm", "-fv", containerName)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerBackend) Pull(image string) {
	args := r.CommonCliArgs
	args = append(args, "pull", image)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerBackend) GetImageInspect(image string) DockerImageInspect {
	args := r.CommonCliArgs
	args = append(args, "image", "inspect", image)
	inspectRaw, err := utils.RunCommand("/", "docker", args...)

	var dockerInspect DockerInspect

	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(inspectRaw, &dockerInspect)).To(Succeed())
	Expect(len(dockerInspect)).To(Equal(1))

	return dockerInspect[0]
}

type DockerInspect []DockerImageInspect
