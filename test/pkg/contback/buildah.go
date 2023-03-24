package contback

import (
	"encoding/json"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/test/pkg/utils"
)

func NewNativeBuildahBackend(isolation thirdparty.Isolation, storageDriver buildah.StorageDriver) ContainerBackend {
	var commonCliArgs []string

	commonBuildahCliArgs, err := buildah.GetBasicBuildahCliArgs(storageDriver)
	Expect(err).NotTo(HaveOccurred())

	commonCliArgs = append(commonCliArgs, commonBuildahCliArgs...)

	return &NativeBuildahBackend{
		BaseContainerBackend: BaseContainerBackend{
			CommonCliArgs: commonCliArgs,
			Isolation:     isolation,
		},
	}
}

type NativeBuildahBackend struct {
	BaseContainerBackend
}

func (r *NativeBuildahBackend) ExpectCmdsToSucceed(image string, cmds ...string) {
	expectCmdsToSucceed(r, image, cmds...)
}

func (r *NativeBuildahBackend) RunSleepingContainer(containerName, image string) {
	args := r.CommonCliArgs
	args = append(args, "from", "--tls-verify=false", "--isolation", r.Isolation.String(), "--format", "docker", "--name", containerName, image)
	utils.RunSucceedCommand("/", "buildah", args...)
}

func (r *NativeBuildahBackend) Exec(containerName string, cmds ...string) {
	for _, cmd := range cmds {
		args := r.CommonCliArgs
		args = append(args, "run", "--isolation", r.Isolation.String(), containerName, "--", "bash", "-o", "pipefail", "-euc", cmd)
		utils.RunSucceedCommand("/", "buildah", args...)
	}
}

func (r *NativeBuildahBackend) Rm(containerName string) {
	args := r.CommonCliArgs
	args = append(args, "rm", containerName)
	utils.RunSucceedCommand("/", "buildah", args...)
}

func (r *NativeBuildahBackend) Pull(image string) {
	args := r.CommonCliArgs
	args = append(args, "pull", "--tls-verify=false", image)
	utils.RunSucceedCommand("/", "buildah", args...)
}

func (r *NativeBuildahBackend) GetImageInspect(image string) DockerImageInspect {
	r.Pull(image)

	args := r.CommonCliArgs
	args = append(args, "inspect", "--type", "image", image)
	inspectRaw, err := utils.RunCommandWithOptions("/", "buildah", args, utils.RunCommandOptions{
		ShouldSucceed: true,
		NoStderr:      true,
	})
	Expect(err).NotTo(HaveOccurred())

	var inspect BuildahInspect
	Expect(json.Unmarshal(inspectRaw, &inspect)).To(Succeed())

	return DockerImageInspect(inspect.Docker)
}

type BuildahInspect struct {
	Docker struct {
		Config       manifest.Schema2Config `json:"config"`
		Architecture string                 `json:"architecture"`
		Os           string                 `json:"os"`
		Variant      string                 `json:"variant"`
	} `json:"Docker"`
}
