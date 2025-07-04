package container_backend

import (
	"bytes"
	"context"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
)

type PerfCheckContainerBackend struct {
	ContainerBackend ContainerBackend
}

func NewPerfCheckContainerBackend(containerBackend ContainerBackend) *PerfCheckContainerBackend {
	return &PerfCheckContainerBackend{ContainerBackend: containerBackend}
}

func (runtime *PerfCheckContainerBackend) Info(ctx context.Context) (info.Info, error) {
	return runtime.ContainerBackend.Info(ctx)
}

func (runtime *PerfCheckContainerBackend) HasStapelBuildSupport() bool {
	return runtime.ContainerBackend.HasStapelBuildSupport()
}

func (runtime *PerfCheckContainerBackend) GetDefaultPlatform() string {
	return runtime.ContainerBackend.GetDefaultPlatform()
}

func (runtime *PerfCheckContainerBackend) GetRuntimePlatform() string {
	return runtime.ContainerBackend.GetRuntimePlatform()
}

func (runtime *PerfCheckContainerBackend) ShouldCleanupDockerfileImage() bool {
	return runtime.ContainerBackend.ShouldCleanupDockerfileImage()
}

func (runtime *PerfCheckContainerBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (resImg *image.Info, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.GetImageInfo %q", ref).
		Do(func() {
			resImg, resErr = runtime.ContainerBackend.GetImageInfo(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Rmi %q", ref).
		Do(func() {
			resErr = runtime.ContainerBackend.Rmi(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Pull(ctx context.Context, ref string, opts PullOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Pull %q", ref).
		Do(func() {
			resErr = runtime.ContainerBackend.Pull(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Tag %q as %q", ref, newRef).
		Do(func() {
			resErr = runtime.ContainerBackend.Tag(ctx, ref, newRef, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Push(ctx context.Context, ref string, opts PushOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Push %q", ref).
		Do(func() {
			resErr = runtime.ContainerBackend.Push(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (resID string, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.BuildDockerfile").
		Do(func() {
			resID, resErr = runtime.ContainerBackend.BuildDockerfile(ctx, dockerfile, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (resID string, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.BuildDockerfile").
		Do(func() {
			resID, resErr = runtime.ContainerBackend.BuildDockerfileStage(ctx, baseImage, opts, instructions...)
		})
	return
}

func (runtime *PerfCheckContainerBackend) BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (resID string, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.BuildDockerfile").
		Do(func() {
			resID, resErr = runtime.ContainerBackend.BuildStapelStage(ctx, baseImage, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) CalculateDependencyImportChecksum(ctx context.Context, dependencyImport DependencyImportSpec, opts CalculateDependencyImportChecksum) (resID string, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.BuildDockerfile").
		Do(func() {
			resID, resErr = runtime.ContainerBackend.CalculateDependencyImportChecksum(ctx, dependencyImport, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.RefreshImageObject %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerBackend.RefreshImageObject(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.PullImageFromRegistry %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerBackend.PullImageFromRegistry(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.RenameImage %q to %q", img.Name(), newImageName).
		Do(func() {
			resErr = runtime.ContainerBackend.RenameImage(ctx, img, newImageName, removeOldName)
		})
	return
}

func (runtime *PerfCheckContainerBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.RemoveImage %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerBackend.RemoveImage(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerBackend) String() string {
	return runtime.ContainerBackend.String()
}

func (runtime *PerfCheckContainerBackend) RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.RemoveHostDirs %q %v", mountDir, dirs).
		Do(func() {
			resErr = runtime.ContainerBackend.RemoveHostDirs(ctx, mountDir, dirs)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Images(ctx context.Context, opts ImagesOptions) (res image.ImagesList, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Images %v", opts).
		Do(func() {
			res, resErr = runtime.ContainerBackend.Images(ctx, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Containers(ctx context.Context, opts ContainersOptions) (res image.ContainerList, resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Containers %v", opts).
		Do(func() {
			res, resErr = runtime.ContainerBackend.Containers(ctx, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) Rm(ctx context.Context, name string, opts RmOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.Rm %q %v", name, opts).
		Do(func() {
			resErr = runtime.ContainerBackend.Rm(ctx, name, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) PostManifest(ctx context.Context, ref string, opts PostManifestOpts) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.PostManifest %q %v", ref, opts).
		Do(func() {
			resErr = runtime.ContainerBackend.PostManifest(ctx, ref, opts)
		})
	return
}

func (runtime *PerfCheckContainerBackend) TagImageByName(ctx context.Context, img LegacyImageInterface) (resErr error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.TagImageByName %q", img.Name()).
		Do(func() {
			resErr = runtime.ContainerBackend.TagImageByName(ctx, img)
		})
	return
}

func (runtime *PerfCheckContainerBackend) ClaimTargetPlatforms(ctx context.Context, targetPlatforms []string) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.ClaimTargetPlatforms %v", targetPlatforms).
		Do(func() { runtime.ContainerBackend.ClaimTargetPlatforms(ctx, targetPlatforms) })
}

func (runtime *PerfCheckContainerBackend) PruneImages(ctx context.Context, options prune.Options) (report prune.Report, err error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.PruneImages %v", options).
		Do(func() {
			report, err = runtime.ContainerBackend.PruneImages(ctx, options)
		})
	return
}

func (runtime *PerfCheckContainerBackend) PruneVolumes(ctx context.Context, options prune.Options) (report prune.Report, err error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.PruneVolumes %v", options).
		Do(func() {
			report, err = runtime.ContainerBackend.PruneVolumes(ctx, options)
		})
	return
}

func (runtime *PerfCheckContainerBackend) DumpImage(ctx context.Context, ref string) (reader *bytes.Reader, err error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.DumpImage %v", ref).
		Do(func() {
			reader, err = runtime.ContainerBackend.DumpImage(ctx, ref)
		})
	return
}

func (runtime *PerfCheckContainerBackend) GenerateSBOM(ctx context.Context, scanOpts scanner.ScanOptions, dstImgLabels []string) (imgId string, err error) {
	logboek.Context(ctx).Default().LogProcess("ContainerBackend.GenerateSBOM scanOpts=%+v, dstImgLabels=%v", scanOpts, dstImgLabels).
		Do(func() {
			imgId, err = runtime.ContainerBackend.GenerateSBOM(ctx, scanOpts, dstImgLabels)
		})
	return
}
