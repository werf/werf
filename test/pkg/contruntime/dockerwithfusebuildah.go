package contruntime

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
)

func NewDockerWithFuseBuildahRuntime() ContainerRuntime {
	return &DockerWithFuseBuildahRuntime{}
}

type DockerWithFuseBuildahRuntime struct {
	BaseContainerRuntime
}

func (r *DockerWithFuseBuildahRuntime) ExpectCmdsToSucceed(image string, cmds ...string) {
	expectCmdsToSucceed(r, image, cmds...)
}

func (r *DockerWithFuseBuildahRuntime) RunSleepingContainer(containerName, image string) {
	args := append(buildahDockerWithFuseDockerArgs(), "from", "--tls-verify=false", "--format", "docker", "--name", containerName, image)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerWithFuseBuildahRuntime) Exec(containerName string, cmds ...string) {
	for _, cmd := range cmds {
		args := append(buildahDockerWithFuseDockerArgs(), "run", containerName, "--", "sh", "-ec", cmd)
		utils.RunSucceedCommand("/", "docker", args...)
	}
}

func (r *DockerWithFuseBuildahRuntime) Rm(containerName string) {
	args := append(buildahDockerWithFuseDockerArgs(), "rm", containerName)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerWithFuseBuildahRuntime) Pull(image string) {
	args := append(buildahDockerWithFuseDockerArgs(), "pull", "--tls-verify=false", image)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerWithFuseBuildahRuntime) GetImageInspectConfig(image string) (config manifest.Schema2Config) {
	r.Pull(image)

	args := append(buildahDockerWithFuseDockerArgs(), "inspect", "--type", "image", image)
	inspectRaw, err := utils.RunCommand("/", "docker", args...)
	Expect(err).NotTo(HaveOccurred())

	var inspect BuildahInspect
	Expect(json.Unmarshal(inspectRaw, &inspect)).To(Succeed())
	return inspect.Docker.Config
}

func buildahDockerWithFuseDockerArgs() []string {
	home, err := os.UserHomeDir()
	Expect(err).NotTo(HaveOccurred())

	args := []string{"run", "--rm"}

	return append(args, buildah.BuildahWithFuseDockerArgs(buildah.BuildahStorageContainerName, filepath.Join(home, ".docker"))...)
}
