package container_backend

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

type DockerServerBackend struct{}

func NewDockerServerBackend() *DockerServerBackend {
	return &DockerServerBackend{}
}

func (backend *DockerServerBackend) ClaimTargetPlatforms(ctx context.Context, targetPlatforms []string) {
	docker.ClaimTargetPlatforms(targetPlatforms)
}

func (backend *DockerServerBackend) GetDefaultPlatform() string {
	return docker.GetDefaultPlatform()
}

func (backend *DockerServerBackend) GetRuntimePlatform() string {
	return docker.GetRuntimePlatform()
}

func (backend *DockerServerBackend) HasStapelBuildSupport() bool {
	return false
}

func (backend *DockerServerBackend) BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (string, error) {
	panic("BuildStapelStage does not implemented for DockerServerBackend. Please report the bug if you've received this message.")
}

func (backend *DockerServerBackend) CalculateDependencyImportChecksum(ctx context.Context, dependencyImport DependencyImportSpec, opts CalculateDependencyImportChecksum) (string, error) {
	panic("CalculateDependencyImportChecksum does not implemented for DockerServerBackend. Please report the bug if you've received this message.")
}

func (backend *DockerServerBackend) BuildDockerfile(ctx context.Context, _ []byte, opts BuildDockerfileOpts) (string, error) {
	switch {
	case opts.BuildContextArchive == nil:
		panic(fmt.Sprintf("BuildContextArchive can't be nil: %+v", opts))
	case opts.DockerfileCtxRelPath == "":
		panic(fmt.Sprintf("DockerfileCtxRelPath can't be empty: %+v", opts))
	}

	var cliArgs []string
	cliArgs = append(cliArgs, "--file", opts.DockerfileCtxRelPath)

	if opts.TargetPlatform != "" {
		cliArgs = append(cliArgs, "--platform", opts.TargetPlatform)
	}
	if opts.Target != "" {
		cliArgs = append(cliArgs, "--target", opts.Target)
	}
	if opts.Network != "" {
		cliArgs = append(cliArgs, "--network", opts.Network)
	}
	if opts.SSH != "" {
		cliArgs = append(cliArgs, "--ssh", opts.SSH)
	}

	for _, addHost := range opts.AddHost {
		cliArgs = append(cliArgs, "--add-host", addHost)
	}
	for _, buildArg := range opts.BuildArgs {
		cliArgs = append(cliArgs, "--build-arg", buildArg)
	}
	for _, label := range opts.Labels {
		cliArgs = append(cliArgs, "--label", label)
	}

	tempID := uuid.New().String()
	opts.Tags = append(opts.Tags, tempID)
	for _, tag := range opts.Tags {
		cliArgs = append(cliArgs, "--tag", tag)
	}

	cliArgs = append(cliArgs, "-")

	if Debug() {
		fmt.Printf("[DOCKER BUILD] docker build %s\n", strings.Join(cliArgs, " "))
	}

	contextReader, err := os.Open(opts.BuildContextArchive.Path())
	if err != nil {
		return "", fmt.Errorf("unable to open context archive %q: %w", opts.BuildContextArchive.Path(), err)
	}
	defer contextReader.Close()

	return tempID, docker.CliBuild_LiveOutputWithCustomIn(ctx, contextReader, cliArgs...)
}

func (backend *DockerServerBackend) BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error) {
	logboek.Context(ctx).Error().LogF("Staged build of Dockerfile is not available for Docker Server backend.")
	logboek.Context(ctx).Error().LogF("Please either:\n")
	logboek.Context(ctx).Error().LogF(" * switch to Buildah backend;\n")
	logboek.Context(ctx).Error().LogF(" * or disable staged build by setting `staged: false` for the image in the werf.yaml.\n")
	logboek.Context(ctx).Error().LogLn()
	return "", fmt.Errorf("staged Dockerfile is not available for Docker Server backend")
}

// ShouldCleanupDockerfileImage for docker-server backend we should cleanup image built from dockerfrom tagged with tempID
// which is implementation detail of the BuildDockerfile.
func (backend *DockerServerBackend) ShouldCleanupDockerfileImage() bool {
	return true
}

func (backend *DockerServerBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	inspect, err := docker.ImageInspect(ctx, ref)
	if client.IsErrNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to inspect docker image: %w", err)
	}
	return docker.NewInfoFromInspect(ref, inspect), nil
}

// GetImageInspect only available for DockerServerBackend
func (backend *DockerServerBackend) GetImageInspect(ctx context.Context, ref string) (*types.ImageInspect, error) {
	inspect, err := docker.ImageInspect(ctx, ref)
	if client.IsErrNotFound(err) {
		return nil, nil
	}
	return inspect, err
}

func (backend *DockerServerBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	if info, err := backend.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}
	return nil
}

// FIXME(multiarch): targetPlatform support needed?
func (backend *DockerServerBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", img.Name(), newImageName)).DoError(func() error {
		if err := docker.CliTag(ctx, img.Name(), newImageName); err != nil {
			return fmt.Errorf("unable to tag image %s by name %s: %w", img.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", img.Name())).DoError(func() error {
			if err := docker.CliRmi(ctx, img.Name()); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	img.SetName(newImageName)

	if info, err := backend.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}

	if desc := img.GetStageDescription(); desc != nil {
		repository, tag := image.ParseRepositoryAndTag(newImageName)
		desc.Info.Name = newImageName
		desc.Info.Repository = repository
		desc.Info.Tag = tag
	}

	return nil
}

func (backend *DockerServerBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) error {
	return logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing image tag %s", img.Name())).DoError(func() error {
		return backend.Rmi(ctx, img.Name(), RmiOpts{
			CommonOpts: CommonOpts{TargetPlatform: img.GetTargetPlatform()},
		})
	})
}

func (backend *DockerServerBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	if err := img.Pull(ctx); err != nil {
		return fmt.Errorf("unable to pull image %s: %w", img.Name(), err)
	}

	if info, err := backend.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %w", img.Name(), err)
	} else {
		img.SetInfo(info)
	}

	return nil
}

func (backend *DockerServerBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	return docker.CliTag(ctx, ref, newRef)
}

func (backend *DockerServerBackend) Push(ctx context.Context, ref string, opts PushOpts) error {
	return docker.CliPushWithRetries(ctx, ref)
}

func (backend *DockerServerBackend) Pull(ctx context.Context, ref string, opts PullOpts) error {
	var args []string
	if opts.TargetPlatform != "" {
		args = append(args, "--platform", opts.TargetPlatform)
	}
	args = append(args, ref)

	if err := docker.CliPull(ctx, args...); err != nil {
		return fmt.Errorf("unable to pull image %s: %w", ref, err)
	}
	return nil
}

func (backend *DockerServerBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	args := []string{ref}
	if opts.Force {
		args = append(args, "--force")
	}
	return docker.CliRmi(ctx, args...)
}

func (backend *DockerServerBackend) Rm(ctx context.Context, ref string, opts RmOpts) error {
	return docker.ContainerRemove(ctx, ref, types.ContainerRemoveOptions{Force: opts.Force})
}

func (backend *DockerServerBackend) PushImage(ctx context.Context, img LegacyImageInterface) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Pushing %s", img.Name())).DoError(func() error {
		return docker.CliPushWithRetries(ctx, img.Name())
	}); err != nil {
		return err
	}
	return nil
}

func (backend *DockerServerBackend) TagImageByName(ctx context.Context, img LegacyImageInterface) error {
	if img.BuiltID() != "" {
		if err := docker.CliTag(ctx, img.BuiltID(), img.Name()); err != nil {
			return fmt.Errorf("unable to tag image %s: %w", img.Name(), err)
		}
	} else {
		if err := backend.RefreshImageObject(ctx, img); err != nil {
			return err
		}
	}
	return nil
}

func (backend *DockerServerBackend) String() string {
	return "docker-server-backend"
}

func (backend *DockerServerBackend) RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) error {
	var containerDirs []string
	for _, dir := range dirs {
		containerDirs = append(containerDirs, util.ToLinuxContainerPath(dir))
	}

	args := []string{
		"--rm",
		"--volume", fmt.Sprintf("%s:%s", mountDir, util.ToLinuxContainerPath(mountDir)),
		"alpine",
		"rm", "-rf",
	}

	args = append(args, containerDirs...)

	return docker.CliRun(ctx, args...)
}

func (backend *DockerServerBackend) Images(ctx context.Context, opts ImagesOptions) (image.ImagesList, error) {
	filterSet := filters.NewArgs()
	for _, item := range opts.Filters {
		filterSet.Add(item.First, item.Second)
	}
	images, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %w", err)
	}

	var res image.ImagesList
	for _, img := range images {
		res = append(res, image.Summary{
			RepoTags: img.RepoTags,
		})
	}
	return res, nil
}

func (backend *DockerServerBackend) Containers(ctx context.Context, opts ContainersOptions) (image.ContainerList, error) {
	filterSet := filters.NewArgs()
	for _, filter := range opts.Filters {
		if filter.ID != "" {
			filterSet.Add("id", filter.ID)
		}
		if filter.Name != "" {
			filterSet.Add("name", filter.Name)
		}
		if filter.Ancestor != "" {
			filterSet.Add("ancestor", filter.Ancestor)
		}
	}

	containersOptions := types.ContainerListOptions{}
	containersOptions.All = true
	containersOptions.Filters = filterSet

	containers, err := docker.Containers(ctx, containersOptions)
	if err != nil {
		return nil, err
	}

	var res image.ContainerList
	for _, container := range containers {
		res = append(res, image.Container{
			ID:      container.ID,
			ImageID: container.ImageID,
			Names:   container.Names,
		})
	}

	return res, nil
}

func (backend *DockerServerBackend) PostManifest(ctx context.Context, ref string, opts PostManifestOpts) error {
	return docker.CreateImage(ctx, ref, docker.CreateImageOptions{Labels: opts.Labels})
}
