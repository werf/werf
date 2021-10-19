package contruntime

import (
	"encoding/json"

	. "github.com/onsi/gomega"
	"github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
)

func NewNativeRootlessBuildahRuntime() ContainerRuntime {
	return &NativeRootlessBuildahRuntime{}
}

type NativeRootlessBuildahRuntime struct {
	BaseContainerRuntime
}

func (r *NativeRootlessBuildahRuntime) ExpectCmdsToSucceed(image string, cmds ...string) {
	expectCmdsToSucceed(r, image, cmds...)
}

func (r *NativeRootlessBuildahRuntime) RunSleepingContainer(containerName, image string) {
	utils.RunSucceedCommand("/",
		"buildah", "from", "--tls-verify=false", "--format", "docker", "--name", containerName, image,
	)
}

func (r *NativeRootlessBuildahRuntime) Exec(containerName string, cmds ...string) {
	for _, cmd := range cmds {
		utils.RunSucceedCommand("/", "buildah", "run", containerName, "--", "sh", "-ec", cmd)
	}
}

func (r *NativeRootlessBuildahRuntime) Rm(containerName string) {
	utils.RunSucceedCommand("/", "buildah", "rm", containerName)
}

func (r *NativeRootlessBuildahRuntime) Pull(image string) {
	utils.RunSucceedCommand("/", "buildah", "pull", "--tls-verify=false", image)
}

func (r *NativeRootlessBuildahRuntime) GetImageInspectConfig(image string) (config manifest.Schema2Config) {
	r.Pull(image)

	inspectRaw, err := utils.RunCommand("/", "buildah", "inspect", "--type", "image", image)
	Expect(err).NotTo(HaveOccurred())
	var inspect BuildahInspect
	Expect(json.Unmarshal(inspectRaw, &inspect)).To(Succeed())

	return inspect.Docker.Config
}
