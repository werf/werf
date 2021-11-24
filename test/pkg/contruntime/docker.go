package contruntime

import (
	"encoding/json"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
)

func NewDockerRuntime() ContainerRuntime {
	return &DockerRuntime{}
}

type DockerRuntime struct {
	BaseContainerRuntime
}

func (r *DockerRuntime) ExpectCmdsToSucceed(image string, cmds ...string) {
	expectCmdsToSucceed(r, image, cmds...)
}

func (r *DockerRuntime) RunSleepingContainer(containerName, image string) {
	utils.RunSucceedCommand("/",
		"docker", "run", "--rm", "-d", "--entrypoint=", "--name", containerName, image, "tail", "-f", "/dev/null",
	)
}

func (r *DockerRuntime) Exec(containerName string, cmds ...string) {
	for _, cmd := range cmds {
		utils.RunSucceedCommand("/", "docker", "exec", containerName, "sh", "-ec", cmd)
	}
}

func (r *DockerRuntime) Rm(containerName string) {
	utils.RunSucceedCommand("/", "docker", "rm", "-fv", containerName)
}

func (r *DockerRuntime) Pull(image string) {
	utils.RunSucceedCommand("/", "docker", "pull", image)
}

func (r *DockerRuntime) GetImageInspectConfig(image string) (config manifest.Schema2Config) {
	configRaw, err := utils.RunCommand("/", "docker", "image", "inspect", "-f", "{{ json .Config }}", image)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(configRaw, &config)).To(Succeed())

	return config
}
