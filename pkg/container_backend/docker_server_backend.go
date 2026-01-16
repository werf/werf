package container_backend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend/info"
	"github.com/werf/werf/v2/pkg/container_backend/prune"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

type DockerServerBackend struct {
	locker lockgate.Locker
}

func NewDockerServerBackend(locker lockgate.Locker) *DockerServerBackend {
	return &DockerServerBackend{
		locker: locker,
	}
}

func (backend *DockerServerBackend) Info(ctx context.Context) (info.Info, error) {
	res := info.Info{}

	sysInfo, err := docker.Info(ctx)
	if err != nil {
		return res, fmt.Errorf("unable to get info: %w", err)
	}

	if sysInfo.OperatingSystem == "Docker Desktop" {
		switch runtime.GOOS {
		case "windows":
			res.StoreGraphRoot = filepath.Join(os.Getenv("HOMEDRIVE"), `\\ProgramData\DockerDesktop\vm-data\`)

		case "darwin":
			res.StoreGraphRoot = filepath.Join(os.Getenv("HOME"), "Library/Containers/com.docker.docker/Data")
		}
	} else {
		res.StoreGraphRoot = sysInfo.DockerRootDir
	}

	return res, nil
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
	} else if opts.SSH == "" && ssh_agent.SSHAuthSock != "" {
		cliArgs = append(cliArgs, "--ssh", "default")
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

	for _, secret := range opts.Secrets {
		cliArgs = append(cliArgs, "--secret", secret)
	}

	for _, tag := range opts.Tags {
		cliArgs = append(cliArgs, "--tag", tag)
	}

	newIDFile, err := tmp_manager.TempFile("docker-built-id-*.tmp")
	if err != nil {
		return "", err
	}
	defer os.Remove(newIDFile.Name())

	cliArgs = append(cliArgs, "--iidfile", newIDFile.Name())

	cliArgs = append(cliArgs, "-")

	if Debug() {
		fmt.Printf("[DOCKER BUILD] docker build %s\n", strings.Join(cliArgs, " "))
	}

	contextReader, err := os.Open(opts.BuildContextArchive.Path())
	if err != nil {
		return "", fmt.Errorf("unable to open context archive %q: %w", opts.BuildContextArchive.Path(), err)
	}
	defer contextReader.Close()

	if err := docker.CliBuild_LiveOutputWithCustomIn(ctx, contextReader, cliArgs...); err != nil {
		return "", err
	}

	newID, err := os.ReadFile(newIDFile.Name())
	if err != nil {
		return "", err
	}

	return string(newID), nil
}

func (backend *DockerServerBackend) BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error) {
	logboek.Context(ctx).Error().LogF("Staged build of Dockerfile is not available for Docker Server backend.\n")
	logboek.Context(ctx).Error().LogF("Please either:\n")
	logboek.Context(ctx).Error().LogF(" * switch to Buildah backend;\n")
	logboek.Context(ctx).Error().LogF(" * or disable staged build by setting `staged: false` for the image in the werf.yaml.\n")
	logboek.Context(ctx).Error().LogLn()
	return "", fmt.Errorf("staged Dockerfile is not available for Docker Server backend")
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

	if stageDesc := img.GetStageDesc(); stageDesc != nil {
		repository, tag := image.ParseRepositoryAndTag(newImageName)
		stageDesc.Info.Name = newImageName
		stageDesc.Info.Repository = repository
		stageDesc.Info.Tag = tag
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
		err = SanitizeError(err)
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
	err := docker.ContainerRemove(ctx, ref, types.ContainerRemoveOptions{Force: opts.Force})
	switch {
	case docker.IsErrContainerPaused(err):
		return errors.Join(ErrCannotRemovePausedContainer, err)
	case docker.IsErrContainerRunning(err):
		return errors.Join(ErrCannotRemoveRunningContainer, err)
	default:
		return err // err or nil
	}
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
		getHostCleanupServiceImage(),
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

	res := make(image.ImagesList, len(images))
	for i, img := range images {
		res[i] = image.Summary{
			ID:          img.ID,
			RepoDigests: img.RepoDigests,
			RepoTags:    img.RepoTags,
			Labels:      img.Labels,
			Created:     time.Unix(img.Created, 0),
			Size:        img.Size,
		}
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

	res := make(image.ContainerList, len(containers))
	for i, container := range containers {
		res[i] = image.Container{
			ID:      container.ID,
			ImageID: container.ImageID,
			Names:   container.Names,
		}
	}

	return res, nil
}

func (backend *DockerServerBackend) PostManifest(ctx context.Context, ref string, opts PostManifestOpts) error {
	return docker.CreateImage(ctx, ref, docker.CreateImageOptions{Labels: opts.Labels})
}

func (backend *DockerServerBackend) PruneImages(ctx context.Context, options prune.Options) (prune.Report, error) {
	report, err := docker.ImagesPrune(ctx, docker.ImagesPruneOptions(options))

	switch {
	case docker.IsErrPruneRunning(err):
		return prune.Report{}, errors.Join(ErrPruneIsAlreadyRunning, err)
	case err != nil:
		return prune.Report{}, err
	}

	return prune.Report(report), err
}

func (backend *DockerServerBackend) PruneVolumes(ctx context.Context, options prune.Options) (prune.Report, error) {
	report, err := docker.VolumesPrune(ctx, docker.VolumesPruneOptions(options))

	switch {
	case docker.IsErrPruneRunning(err):
		return prune.Report{}, errors.Join(ErrPruneIsAlreadyRunning, err)
	case err != nil:
		return prune.Report{}, err
	}

	return prune.Report(report), err
}

func (backend *DockerServerBackend) SaveImageToStream(ctx context.Context, imageName string) (io.ReadCloser, error) {
	return docker.CliImageSaveToStream(ctx, imageName)
}

func (backend *DockerServerBackend) LoadImageFromStream(ctx context.Context, input io.Reader) (string, error) {
	return docker.CliLoadFromStream(ctx, input)
}
