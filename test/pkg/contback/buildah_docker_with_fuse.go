package contback

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/test/pkg/utils"
)

func NewDockerWithFuseBuildahBackend(isolation thirdparty.Isolation, storageDriver buildah.StorageDriver) ContainerBackend {
	home, err := os.UserHomeDir()
	Expect(err).NotTo(HaveOccurred())

	commonCliArgs := append([]string{"run", "--rm"}, buildah.BuildahWithFuseDockerArgs(buildah.BuildahStorageContainerName, filepath.Join(home, ".docker"))...)

	commonBuildahCliArgs, err := buildah.GetBasicBuildahCliArgs(storageDriver)
	Expect(err).NotTo(HaveOccurred())

	commonCliArgs = append(commonCliArgs, commonBuildahCliArgs...)

	return &DockerWithFuseBuildahBackend{
		BaseContainerBackend: BaseContainerBackend{
			CommonCliArgs: commonCliArgs,
			Isolation:     isolation,
		},
	}
}

type DockerWithFuseBuildahBackend struct {
	BaseContainerBackend
}

func (r *DockerWithFuseBuildahBackend) ExpectCmdsToSucceed(image string, cmds ...string) {
	expectCmdsToSucceed(r, image, cmds...)
}

func (r *DockerWithFuseBuildahBackend) RunSleepingContainer(containerName, image string) {
	args := r.CommonCliArgs
	args = append(args, "from", "--tls-verify=false", "--isolation", r.Isolation.String(), "--format", "docker", "--name", containerName, image)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerWithFuseBuildahBackend) Exec(containerName string, cmds ...string) {
	for _, cmd := range cmds {
		args := r.CommonCliArgs
		args = append(args, "run", "--isolation", r.Isolation.String(), containerName, "--", "sh", "-ec", cmd)
		utils.RunSucceedCommand("/", "docker", args...)
	}
}

func (r *DockerWithFuseBuildahBackend) Rm(containerName string) {
	args := r.CommonCliArgs
	args = append(args, "rm", containerName)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerWithFuseBuildahBackend) Pull(image string) {
	args := r.CommonCliArgs
	args = append(args, "pull", "--tls-verify=false", image)
	utils.RunSucceedCommand("/", "docker", args...)
}

func (r *DockerWithFuseBuildahBackend) GetImageInspectConfig(image string) (config manifest.Schema2Config) {
	r.Pull(image)

	args := r.CommonCliArgs
	args = append(args, "inspect", "--type", "image", image)
	inspectRaw, err := utils.RunCommand("/", "docker", args...)
	Expect(err).NotTo(HaveOccurred())

	var inspect BuildahInspect
	Expect(json.Unmarshal(inspectRaw, &inspect)).To(Succeed())
	return inspect.Docker.Config
}
