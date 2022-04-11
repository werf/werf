package container_backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/otiai10/copy"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

type BuildahBackend struct {
	buildah buildah.Buildah
}

type BuildahImage struct {
	Image LegacyImageInterface
}

func NewBuildahBackend(buildah buildah.Buildah) *BuildahBackend {
	return &BuildahBackend{buildah: buildah}
}

func (runtime *BuildahBackend) HasStapelBuildSupport() bool {
	return true
}

func (runtime *BuildahBackend) getBuildahCommonOpts(ctx context.Context, suppressLog bool) (opts buildah.CommonOpts) {
	if !suppressLog {
		opts.LogWriter = logboek.Context(ctx).OutStream()
	}

	return
}

type containerDesc struct {
	ImageName string
	Name      string
	RootMount string
}

func (runtime *BuildahBackend) createContainers(ctx context.Context, images []string) ([]*containerDesc, error) {
	var res []*containerDesc

	for _, img := range images {
		containerID := fmt.Sprintf("werf-stage-build-%s", uuid.New().String())

		_, err := runtime.buildah.FromCommand(ctx, containerID, img, buildah.FromCommandOpts(runtime.getBuildahCommonOpts(ctx, true)))
		if err != nil {
			return nil, fmt.Errorf("unable to create container using base image %q: %w", img, err)
		}

		res = append(res, &containerDesc{ImageName: img, Name: containerID})
	}

	return res, nil
}

func (runtime *BuildahBackend) removeContainers(ctx context.Context, containers []*containerDesc) error {
	for _, cont := range containers {
		if err := runtime.buildah.Rm(ctx, cont.Name, buildah.RmOpts(runtime.getBuildahCommonOpts(ctx, true))); err != nil {
			return fmt.Errorf("unable to remove container %q: %w", cont.Name, err)
		}
	}

	return nil
}

func (runtime *BuildahBackend) mountContainers(ctx context.Context, containers []*containerDesc) error {
	for _, cont := range containers {
		containerRoot, err := runtime.buildah.Mount(ctx, cont.Name, buildah.MountOpts(runtime.getBuildahCommonOpts(ctx, true)))
		if err != nil {
			return fmt.Errorf("unable to mount container %q root dir: %w", cont.Name, err)
		}
		cont.RootMount = containerRoot
	}

	return nil
}

func (runtime *BuildahBackend) unmountContainers(ctx context.Context, containers []*containerDesc) error {
	for _, cont := range containers {
		if err := runtime.buildah.Umount(ctx, cont.Name, buildah.UmountOpts(runtime.getBuildahCommonOpts(ctx, true))); err != nil {
			return fmt.Errorf("container %q: %w", cont.Name, err)
		}
	}

	return nil
}

func (runtime *BuildahBackend) applyCommands(ctx context.Context, container *containerDesc, buildVolumes []string, commands []string) error {
	for _, cmd := range commands {
		var mounts []specs.Mount
		mounts, err := makeBuildahMounts(buildVolumes)
		if err != nil {
			return err
		}

		// TODO(stapel-to-buildah): Consider support for shell script instead of separate run commands to allow shared
		// 							  usage of shell variables and functions between multiple commands.
		//                          Maybe there is no need of such function, instead provide options to select shell in the werf.yaml.
		//                          Is it important to provide compatibility between docker-server-based werf.yaml and buildah-based?
		if err := runtime.buildah.RunCommand(ctx, container.Name, []string{"sh", "-c", cmd}, buildah.RunCommandOpts{
			CommonOpts: runtime.getBuildahCommonOpts(ctx, false),
			Mounts:     mounts,
		}); err != nil {
			return fmt.Errorf("unable to run %q: %w", cmd, err)
		}
	}

	return nil
}

type dependencyContainer struct {
	Container *containerDesc
	Import    DependencyImport
}

func (runtime *BuildahBackend) applyDataArchives(ctx context.Context, container *containerDesc, dataArchives []DataArchive) error {
	for _, archive := range dataArchives {
		destPath := filepath.Join(container.RootMount, archive.To)

		var extractDestPath string
		switch archive.Type {
		case DirectoryArchive:
			extractDestPath = destPath
		case FileArchive:
			extractDestPath = filepath.Dir(destPath)
		default:
			return fmt.Errorf("unknown archive type %q", archive.Type)
		}

		_, err := os.Stat(destPath)
		switch {
		case os.IsNotExist(err):
		case err != nil:
			return fmt.Errorf("unable to access container path %q: %w", destPath, err)
		default:
			logboek.Context(ctx).Debug().LogF("Removing archive destination path %s\n", archive.To)
			if err := os.RemoveAll(destPath); err != nil {
				return fmt.Errorf("unable to cleanup archive destination path %s: %w", archive.To, err)
			}
		}

		logboek.Context(ctx).Debug().LogF("Extracting archive into container path %s\n", archive.To)
		if err := util.ExtractTar(archive.Data, extractDestPath); err != nil {
			return fmt.Errorf("unable to extract data archive into %s: %w", archive.To, err)
		}
		if err := archive.Data.Close(); err != nil {
			return fmt.Errorf("error closing archive data stream: %w", err)
		}
	}

	return nil
}

func (runtime *BuildahBackend) applyPathsToRemove(ctx context.Context, container *containerDesc, pathsToRemove []string) error {
	for _, path := range pathsToRemove {
		destPath := filepath.Join(container.RootMount, path)

		logboek.Context(ctx).Debug().LogF("Removing container path %s\n", path)
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("unable to remove path %s: %w", path, err)
		}
	}

	return nil
}

func (runtime *BuildahBackend) applyDependenciesImports(ctx context.Context, container *containerDesc, dependenciesImports []DependencyImport) error {
	var dependencies []*dependencyContainer

	var dependenciesImages []string
	for _, imp := range dependenciesImports {
		dependenciesImages = append(dependenciesImages, imp.ImageName)
	}

	logboek.Context(ctx).Debug().LogF("Creating containers for dependencies images %v\n", dependenciesImages)
	dependenciesContainers, err := runtime.createContainers(ctx, dependenciesImages)
	if err != nil {
		return fmt.Errorf("unable to create dependencies containers: %w", err)
	}
	defer func() {
		if err := runtime.removeContainers(ctx, dependenciesContainers); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporal dependencies containers: %s\n", err)
		}
	}()

	for _, cont := range dependenciesContainers {
	FindImport:
		for _, imp := range dependenciesImports {
			if imp.ImageName == cont.ImageName {
				dependencies = append(dependencies, &dependencyContainer{
					Container: cont,
					Import:    imp,
				})

				break FindImport
			}
		}
	}

	// NOTE: maybe it is more optimal not to mount all dependencies at once, but mount one-by-one
	logboek.Context(ctx).Debug().LogF("Mounting dependencies containers %v\n", dependenciesContainers)
	if err := runtime.mountContainers(ctx, dependenciesContainers); err != nil {
		return fmt.Errorf("unable to mount containers: %w", err)
	}
	defer func() {
		logboek.Context(ctx).Debug().LogF("Unmounting dependencies containers %v\n", dependenciesContainers)
		if err := runtime.unmountContainers(ctx, dependenciesContainers); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to unmount containers: %s\n", err)
		}
	}()

	for _, dep := range dependencies {
		copyFrom := filepath.Join(dep.Container.RootMount, dep.Import.FromPath)
		copyTo := filepath.Join(container.RootMount, dep.Import.ToPath)
		fmt.Printf("copying from %q to %q\n", copyFrom, copyTo)
		logboek.Context(ctx).Debug().LogF("Copying dependency %v from %q to %q\n", dep.Import, copyFrom, copyTo)
		if err := copy.Copy(copyFrom, copyTo); err != nil {
			return fmt.Errorf("unable to copy %s to %s: %w", copyFrom, copyTo, err)
		}
	}

	return nil
}

func (runtime *BuildahBackend) BuildStapelStage(ctx context.Context, opts BuildStapelStageOptions) (string, error) {
	var container *containerDesc
	if c, err := runtime.createContainers(ctx, []string{opts.BaseImage}); err != nil {
		return "", err
	} else {
		container = c[0]
	}
	defer func() {
		if err := runtime.removeContainers(ctx, []*containerDesc{container}); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporal build container: %s\n", err)
		}
	}()
	// TODO(stapel-to-buildah): cleanup orphan build containers in werf-host-cleanup procedure

	if len(opts.DependenciesImports)+len(opts.DataArchives)+len(opts.PathsToRemove) > 0 {
		logboek.Context(ctx).Debug().LogF("Mounting build container %s\n", container.Name)
		if err := runtime.mountContainers(ctx, []*containerDesc{container}); err != nil {
			return "", fmt.Errorf("unable to mount build container %s: %w", container.Name, err)
		}
		defer func() {
			logboek.Context(ctx).Debug().LogF("Unmounting build container %s\n", container.Name)
			if err := runtime.unmountContainers(ctx, []*containerDesc{container}); err != nil {
				logboek.Context(ctx).Error().LogF("ERROR: unable to unmount containers: %s\n", err)
			}
		}()
	}

	if len(opts.DependenciesImports) > 0 {
		if err := runtime.applyDependenciesImports(ctx, container, opts.DependenciesImports); err != nil {
			return "", err
		}
	}
	if len(opts.DataArchives) > 0 {
		if err := runtime.applyDataArchives(ctx, container, opts.DataArchives); err != nil {
			return "", err
		}
	}
	if len(opts.PathsToRemove) > 0 {
		if err := runtime.applyPathsToRemove(ctx, container, opts.PathsToRemove); err != nil {
			return "", err
		}
	}
	if len(opts.Commands) > 0 {
		if err := runtime.applyCommands(ctx, container, opts.BuildVolumes, opts.Commands); err != nil {
			return "", err
		}
	}

	logboek.Context(ctx).Debug().LogF("Setting config for build container %q\n", container.Name)
	if err := runtime.buildah.Config(ctx, container.Name, buildah.ConfigOpts{
		CommonOpts:  runtime.getBuildahCommonOpts(ctx, true),
		Labels:      opts.Labels,
		Volumes:     opts.Volumes,
		Expose:      opts.Expose,
		Envs:        opts.Envs,
		Cmd:         opts.Cmd,
		Entrypoint:  opts.Entrypoint,
		User:        opts.User,
		Workdir:     opts.Workdir,
		Healthcheck: opts.Healthcheck,
	}); err != nil {
		return "", fmt.Errorf("unable to set container %q config: %w", container.Name, err)
	}

	// TODO(stapel-to-buildah): Save container name as builtID. There is no need to commit an image here,
	//                            because buildah allows to commit and push directly container, which would happen later.
	logboek.Context(ctx).Debug().LogF("committing container %q\n", container.Name)
	imgID, err := runtime.buildah.Commit(ctx, container.Name, buildah.CommitOpts{CommonOpts: runtime.getBuildahCommonOpts(ctx, true)})
	if err != nil {
		return "", fmt.Errorf("unable to commit container %q: %w", container.Name, err)
	}

	return imgID, nil
}

// GetImageInfo returns nil, nil if image not found.
func (runtime *BuildahBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	inspect, err := runtime.buildah.Inspect(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("error getting buildah inspect of %q: %w", ref, err)
	}
	if inspect == nil {
		return nil, nil
	}

	repository, tag := image.ParseRepositoryAndTag(ref)

	return &image.Info{
		Name:              ref,
		Repository:        repository,
		Tag:               tag,
		Labels:            inspect.Docker.Config.Labels,
		CreatedAtUnixNano: inspect.Docker.Created.UnixNano(),
		// RepoDigest:        repoDigest, // FIXME
		OnBuild:  inspect.Docker.Config.OnBuild,
		ID:       inspect.Docker.ID,
		ParentID: inspect.Docker.Config.Image,
		Size:     inspect.Docker.Size,
	}, nil
}

func (runtime *BuildahBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	return runtime.buildah.Rmi(ctx, ref, buildah.RmiOpts{
		Force: true,
		CommonOpts: buildah.CommonOpts{
			LogWriter: logboek.Context(ctx).OutStream(),
		},
	})
}

func (runtime *BuildahBackend) Pull(ctx context.Context, ref string, opts PullOpts) error {
	return runtime.buildah.Pull(ctx, ref, buildah.PullOpts{
		LogWriter: logboek.Context(ctx).OutStream(),
	})
}

func (runtime *BuildahBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	return runtime.buildah.Tag(ctx, ref, newRef, buildah.TagOpts{
		LogWriter: logboek.Context(ctx).OutStream(),
	})
}

func (runtime *BuildahBackend) Push(ctx context.Context, ref string, opts PushOpts) error {
	return runtime.buildah.Push(ctx, ref, buildah.PushOpts{
		LogWriter: logboek.Context(ctx).OutStream(),
	})
}

func (runtime *BuildahBackend) BuildDockerfile(ctx context.Context, dockerfile []byte, opts BuildDockerfileOpts) (string, error) {
	buildArgs := make(map[string]string)
	for _, argStr := range opts.BuildArgs {
		argParts := strings.SplitN(argStr, "=", 2)
		if len(argParts) < 2 {
			return "", fmt.Errorf("invalid build argument %q given, expected string in the key=value format", argStr)
		}
		buildArgs[argParts[0]] = argParts[1]
	}

	return runtime.buildah.BuildFromDockerfile(ctx, dockerfile, buildah.BuildFromDockerfileOpts{
		CommonOpts: buildah.CommonOpts{
			LogWriter: logboek.Context(ctx).OutStream(),
		},
		ContextTar: opts.ContextTar,
		BuildArgs:  buildArgs,
		Target:     opts.Target,
	})
}

func (runtime *BuildahBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}
	return nil
}

func (runtime *BuildahBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	if err := runtime.Pull(ctx, img.Name(), PullOpts{}); err != nil {
		return fmt.Errorf("unable to pull image %s: %w", img.Name(), err)
	}

	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %w", img.Name(), err)
	} else {
		img.SetInfo(info)
	}

	return nil
}

func (runtime *BuildahBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", img.Name(), newImageName)).DoError(func() error {
		if err := runtime.Tag(ctx, img.Name(), newImageName, TagOpts{}); err != nil {
			return fmt.Errorf("unable to tag image %s by name %s: %w", img.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", img.Name())).DoError(func() error {
			if err := runtime.Rmi(ctx, img.Name(), RmiOpts{}); err != nil {
				return fmt.Errorf("unable to remove image %q: %w", img.Name(), err)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	img.SetName(newImageName)

	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}

	desc := img.GetStageDescription()

	repository, tag := image.ParseRepositoryAndTag(newImageName)
	desc.Info.Repository = repository
	desc.Info.Tag = tag

	return nil
}

func (runtime *BuildahBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing image tag %s", img.Name())).DoError(func() error {
		if err := runtime.Rmi(ctx, img.Name(), RmiOpts{}); err != nil {
			return fmt.Errorf("unable to remove image %q: %w", img.Name(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (runtime *BuildahBackend) String() string {
	return "buildah-runtime"
}

func parseVolume(volume string) (string, string, error) {
	volumeParts := strings.SplitN(volume, ":", 2)
	if len(volumeParts) != 2 {
		return "", "", fmt.Errorf("expected SOURCE:DESTINATION format")
	}
	return volumeParts[0], volumeParts[1], nil
}

func makeBuildahMounts(volumes []string) ([]specs.Mount, error) {
	var mounts []specs.Mount

	for _, volume := range volumes {
		from, to, err := parseVolume(volume)
		if err != nil {
			return nil, fmt.Errorf("invalid volume %q: %w", volume, err)
		}

		mounts = append(mounts, specs.Mount{
			Type:        "bind",
			Source:      from,
			Destination: to,
		})
	}

	return mounts, nil
}
