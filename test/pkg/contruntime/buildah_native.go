package contruntime

import (
	"encoding/json"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/test/pkg/utils"
)

func NewNativeBuildahRuntime(isolation thirdparty.Isolation, storageDriver buildah.StorageDriver) ContainerRuntime {
	var commonCliArgs []string

	commonBuildahCliArgs, err := buildah.GetBasicBuildahCliArgs(storageDriver)
	Expect(err).NotTo(HaveOccurred())

	commonCliArgs = append(commonCliArgs, commonBuildahCliArgs...)

	return &NativeBuildahRuntime{
		BaseContainerRuntime: BaseContainerRuntime{
			CommonCliArgs: commonCliArgs,
			Isolation:     isolation,
		},
	}
}

type NativeBuildahRuntime struct {
	BaseContainerRuntime
}

func (r *NativeBuildahRuntime) ExpectCmdsToSucceed(image string, cmds ...string) {
	expectCmdsToSucceed(r, image, cmds...)
}

func (r *NativeBuildahRuntime) RunSleepingContainer(containerName, image string) {
	args := r.CommonCliArgs
	args = append(args, "from", "--tls-verify=false", "--isolation", r.Isolation.String(), "--format", "docker", "--name", containerName, image)
	utils.RunSucceedCommand("/", "buildah", args...)
}

func (r *NativeBuildahRuntime) Exec(containerName string, cmds ...string) {
	for _, cmd := range cmds {
		args := r.CommonCliArgs
		args = append(args, "run", "--isolation", r.Isolation.String(), containerName, "--", "sh", "-ec", cmd)
		utils.RunSucceedCommand("/", "buildah", args...)
	}
}

func (r *NativeBuildahRuntime) Rm(containerName string) {
	args := r.CommonCliArgs
	args = append(args, "rm", containerName)
	utils.RunSucceedCommand("/", "buildah", args...)
}

func (r *NativeBuildahRuntime) Pull(image string) {
	args := r.CommonCliArgs
	args = append(args, "pull", "--tls-verify=false", image)
	utils.RunSucceedCommand("/", "buildah", args...)
}

func (r *NativeBuildahRuntime) GetImageInspectConfig(image string) (config manifest.Schema2Config) {
	r.Pull(image)

	args := r.CommonCliArgs
	args = append(args, "inspect", "--type", "image", image)
	inspectRaw, err := utils.RunCommand("/", "buildah", args...)
	Expect(err).NotTo(HaveOccurred())

	var inspect BuildahInspect
	Expect(json.Unmarshal(inspectRaw, &inspect)).To(Succeed())

	return inspect.Docker.Config
}
