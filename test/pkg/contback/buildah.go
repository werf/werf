package contback

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/buildah/thirdparty"
	"github.com/werf/werf/v2/test/pkg/thirdparty/contruntime/manifest"
	"github.com/werf/werf/v2/test/pkg/utils"
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

func (r *NativeBuildahBackend) ExpectCmdsToSucceed(ctx context.Context, image string, cmds ...string) {
	expectCmdsToSucceed(ctx, r, image, cmds...)
}

func (r *NativeBuildahBackend) RunSleepingContainer(ctx context.Context, containerName, image string) {
	args := r.CommonCliArgs
	args = append(args, "from", "--tls-verify=false", "--isolation", r.Isolation.String(), "--format", "docker", "--name", containerName, image)
	utils.RunSucceedCommand(ctx, "/", "buildah", args...)
}

func (r *NativeBuildahBackend) Exec(ctx context.Context, containerName string, cmds ...string) {
	for _, cmd := range cmds {
		args := r.CommonCliArgs
		args = append(args, "run", "--isolation", r.Isolation.String(), containerName, "--", "bash", "-o", "pipefail", "-euc", cmd)
		utils.RunSucceedCommand(ctx, "/", "buildah", args...)
	}
}

func (r *NativeBuildahBackend) Rm(ctx context.Context, containerName string) {
	args := r.CommonCliArgs
	args = append(args, "rm", containerName)
	utils.RunSucceedCommand(ctx, "/", "buildah", args...)
}

func (r *NativeBuildahBackend) Pull(ctx context.Context, image string) {
	args := r.CommonCliArgs
	args = append(args, "pull", "--tls-verify=false", image)
	utils.RunSucceedCommand(ctx, "/", "buildah", args...)
}

func (r *NativeBuildahBackend) SaveImageToStream(ctx context.Context, image string) io.ReadCloser {
	// Buildah doesn't support redirecting to stdout
	// https://github.com/containers/buildah/issues/936
	// So we should use tmp file
	tmpDir, err := os.MkdirTemp(os.TempDir(), "buildah-img-*")
	Expect(err).NotTo(HaveOccurred())

	// NOTE: Command "buildah push <image_ref> oci-archive:/path/to/dir" doesn't create manifest.json file.
	// We must use "dir:/" transport to create manifest.json file.
	// 1. Push image to tmp dir
	args := r.CommonCliArgs
	args = append(args, "push", "--format", "v2s2", "--disable-compression", image, fmt.Sprintf("dir:%s", tmpDir))
	utils.RunSucceedCommand(ctx, "/", "buildah", args...)

	// 2. Create tar archive from tmp dir AND redirect result to stdout.
	// Option --transform converts ./manifest.json to manifest.json.
	argsTar := []string{"-c", "-f", "-", "--transform", "s|^\\./||", "."}
	b, err := utils.RunCommandWithOptions(ctx, tmpDir, "tar", argsTar, utils.RunCommandOptions{
		NoStderr:      true,
		ShouldSucceed: true,
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(os.RemoveAll(tmpDir)).To(Succeed())

	return io.NopCloser(bytes.NewReader(b))
}

func (r *NativeBuildahBackend) GetImageInspect(ctx context.Context, image string) DockerImageInspect {
	r.Pull(ctx, image)

	args := r.CommonCliArgs
	args = append(args, "inspect", "--type", "image", image)
	inspectRaw, err := utils.RunCommandWithOptions(ctx, "/", "buildah", args, utils.RunCommandOptions{
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
		Author       string                 `json:"author"`
		Config       manifest.Schema2Config `json:"config"`
		Architecture string                 `json:"architecture"`
		Os           string                 `json:"os"`
		Variant      string                 `json:"variant"`
		History      interface{}            `json:"history"`
	} `json:"Docker"`
}
