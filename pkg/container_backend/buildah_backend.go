package container_backend

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/otiai10/copy"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/path_matcher"
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

func (runtime *BuildahBackend) applyDataArchives(ctx context.Context, container *containerDesc, dataArchives []DataArchiveSpec) error {
	for _, archive := range dataArchives {
		destPath := filepath.Join(container.RootMount, archive.To)

		var extractDestPath string
		switch archive.Type {
		case DirectoryArchive:
			extractDestPath = destPath
		case FileArchive:
			extractDestPath = filepath.Dir(destPath)

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
		default:
			return fmt.Errorf("unknown archive type %q", archive.Type)
		}

		logboek.Context(ctx).Debug().LogF("Apply data archive into %q\n", archive.To)

		logboek.Context(ctx).Debug().LogF("Extracting archive into container path %s\n", archive.To)
		if err := util.ExtractTar(archive.Archive, extractDestPath); err != nil {
			return fmt.Errorf("unable to extract data archive into %s: %w", archive.To, err)
		}
		if err := archive.Archive.Close(); err != nil {
			return fmt.Errorf("error closing archive data stream: %w", err)
		}
	}

	return nil
}

func (runtime *BuildahBackend) applyRemoveData(ctx context.Context, container *containerDesc, removeData []RemoveDataSpec) error {
	for _, spec := range removeData {
		switch spec.Type {
		case RemoveExactPath:
			for _, path := range spec.Paths {
				destPath := filepath.Join(container.RootMount, path)
				if err := removeExactPath(ctx, destPath); err != nil {
					return fmt.Errorf("unable to remove %q: %w", path, err)
				}
			}
		case RemoveExactPathWithEmptyParentDirs:
			for _, path := range spec.Paths {
				destPath := filepath.Join(container.RootMount, path)
				if err := removeExactPathWithEmptyParentDirs(ctx, destPath, spec.KeepParentDirs); err != nil {
					return fmt.Errorf("unable to remove %q: %w", path, err)
				}
			}
		case RemoveInsidePath:
			for _, path := range spec.Paths {
				destPath := filepath.Join(container.RootMount, path)
				if err := removeInsidePath(ctx, destPath); err != nil {
					return fmt.Errorf("unable to remove %q: %w", path, err)
				}
			}
		default:
			return fmt.Errorf("unknown remove operation type %q", spec.Type)
		}
	}

	return nil
}

func (runtime *BuildahBackend) applyDependenciesImports(ctx context.Context, container *containerDesc, depImports []DependencyImportSpec) error {
	var depImages []string
	for _, imp := range depImports {
		if util.IsStringsContainValue(depImages, imp.ImageName) {
			continue
		}

		depImages = append(depImages, imp.ImageName)
	}

	logboek.Context(ctx).Debug().LogF("Creating containers for depContainers images %v\n", depImages)
	createdDepContainers, err := runtime.createContainers(ctx, depImages)
	if err != nil {
		return fmt.Errorf("unable to create depContainers containers: %w", err)
	}
	defer func() {
		if err := runtime.removeContainers(ctx, createdDepContainers); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporal depContainers containers: %s\n", err)
		}
	}()

	// NOTE: maybe it is more optimal not to mount all dependencies at once, but mount one-by-one
	logboek.Context(ctx).Debug().LogF("Mounting depContainers containers %v\n", createdDepContainers)
	if err := runtime.mountContainers(ctx, createdDepContainers); err != nil {
		return fmt.Errorf("unable to mount containers: %w", err)
	}
	defer func() {
		logboek.Context(ctx).Debug().LogF("Unmounting depContainers containers %v\n", createdDepContainers)
		if err := runtime.unmountContainers(ctx, createdDepContainers); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to unmount containers: %s\n", err)
		}
	}()

	for _, dep := range createdDepContainers {
		for _, imp := range depImports {
			if imp.ImageName != dep.ImageName {
				continue
			}

			absFrom := filepath.Join(dep.RootMount, imp.FromPath)
			absTo := filepath.Join(container.RootMount, imp.ToPath)

			var uid, gid *uint32
			if uid, gid, err = getUIDAndGID(imp.Owner, imp.Group, container.RootMount); err != nil {
				return fmt.Errorf("error getting UID/GID: %w", err)
			}

			pathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
				IncludeGlobs: imp.IncludePaths,
				ExcludeGlobs: imp.ExcludePaths,
			})

			if err := copyFromTo(ctx, uid, gid, absFrom, absTo, pathMatcher); err != nil {
				return fmt.Errorf("error copying dependency import files from %q to %q: %w", absFrom, absTo, err)
			}

			if err := updateOwner(absTo, uid, gid); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("error updating file ownersnip: %w", err)
			}
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

	if len(opts.DependencyImportSpecs)+len(opts.DataArchiveSpecs)+len(opts.RemoveDataSpecs) > 0 {
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

	if len(opts.DependencyImportSpecs) > 0 {
		if err := runtime.applyDependenciesImports(ctx, container, opts.DependencyImportSpecs); err != nil {
			return "", err
		}
	}
	if len(opts.DataArchiveSpecs) > 0 {
		if err := runtime.applyDataArchives(ctx, container, opts.DataArchiveSpecs); err != nil {
			return "", err
		}
	}
	if len(opts.RemoveDataSpecs) > 0 {
		if err := runtime.applyRemoveData(ctx, container, opts.RemoveDataSpecs); err != nil {
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

func copyFromTo(ctx context.Context, uid *uint32, gid *uint32, absFrom, absTo string, pathMatcher path_matcher.PathMatcher) error {
	if err := walkPath(absFrom, func(rootPathIsFile bool, relSrc string, dirEntry *fs.DirEntry, err error) error {
		if rootPathIsFile {
			logboek.Context(ctx).Debug().LogF("Copying dependency import %q to %q\n", absFrom, absTo)
			if err := copy.Copy(absFrom, absTo); err != nil {
				return fmt.Errorf("error copying file %q to %q: %w", absFrom, absTo, err)
			}
			return nil
		}

		if err != nil {
			return fmt.Errorf("error while walking dir %q: %w", absFrom, err)
		}

		absSrc := filepath.Join(absFrom, relSrc)
		absDst := filepath.Join(absTo, relSrc)

		srcFileInfo, err := (*dirEntry).Info()
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("file %q was renamed or removed while we were trying to get info about it", absSrc)
		} else if err != nil {
			return fmt.Errorf("error getting info about file %q", absSrc)
		}

		if srcFileInfo.IsDir() {
			if !pathMatcher.IsDirOrSubmodulePathMatched(relSrc) {
				return fs.SkipDir
			}

			if pathMatcher.ShouldGoThrough(relSrc) {
				if err := os.MkdirAll(relSrc, srcFileInfo.Mode()); err != nil {
					return fmt.Errorf("error creating directory %q: %w", absSrc, err)
				}
				return nil
			}
		} else if !pathMatcher.IsPathMatched(relSrc) {
			return nil
		}

		logboek.Context(ctx).Debug().LogF("Copying dependency import %q to %q\n", absSrc, absDst)
		if err := copy.Copy(absSrc, absDst); err != nil {
			return fmt.Errorf("error copying file %q to %q: %w", absSrc, absDst, err)
		}

		if err := updateOwner(absDst, uid, gid); err != nil {
			return fmt.Errorf("error updating file ownership: %w", err)
		}

		if srcFileInfo.IsDir() {
			return fs.SkipDir
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error walking dependency import source root: %w", err)
	}
	return nil
}

func walkPath(path string, fn func(rootPathIsFile bool, entryRelPath string, dirEntry *fs.DirEntry, err error) error) error {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("error getting file info for path %q: %w", path, err)
	}

	if !fileInfo.IsDir() {
		return fn(true, "", nil, nil)
	} else {
		rootFs := os.DirFS(path)
		if err := fs.WalkDir(rootFs, ".", func(relSrc string, entry fs.DirEntry, err error) error {
			return fn(false, relSrc, &entry, err)
		}); err != nil {
			return fmt.Errorf("error walking directory %q: %w", rootFs, err)
		}
		return nil
	}
}

func updateOwner(path string, uid *uint32, gid *uint32) error {
	if uid == nil && gid == nil {
		return nil
	}

	if err := walkPath(path, func(rootPathIsFile bool, entryRelPath string, dirEntry *fs.DirEntry, err error) error {
		if rootPathIsFile {
			return os.Lchown(path, uint32PtrUIDOrGIDToInt(uid), uint32PtrUIDOrGIDToInt(gid))
		}

		return os.Lchown(filepath.Join(path, entryRelPath), uint32PtrUIDOrGIDToInt(uid), uint32PtrUIDOrGIDToInt(gid))
	}); err != nil {
		return fmt.Errorf("error walking path %q: %w", path, err)
	}

	return nil
}

func uint32PtrUIDOrGIDToInt(uidOrGid *uint32) int {
	if uidOrGid == nil {
		return -1
	}

	return int(*uidOrGid)
}

func getUIDAndGID(userNameOrUID, groupNameOrGID, fsRoot string) (*uint32, *uint32, error) {
	uid, err := getUID(userNameOrUID, fsRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting UID: %w", err)
	}

	gid, err := getGID(groupNameOrGID, fsRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting GID: %w", err)
	}

	return uid, gid, nil
}

// Returns nil pointer if username/UID is empty string.
func getUID(userNameOrUID string, fsRoot string) (*uint32, error) {
	var uid *uint32
	if userNameOrUID != "" {
		if parsed, err := strconv.ParseUint(userNameOrUID, 10, 32); errors.Is(err, strconv.ErrSyntax) {
			result, err := getUIDFromUserName(userNameOrUID, filepath.Join(fsRoot, "etc", "passwd"))
			if err != nil {
				return nil, fmt.Errorf("error getting UID from user name: %w", err)
			}
			uid = &result
		} else if err != nil {
			return nil, fmt.Errorf("error parsing user ID: %w", err)
		} else {
			result := uint32(parsed)
			uid = &result
		}
	}
	return uid, nil
}

// Returns nil pointer if groupname/GID is empty string.
func getGID(groupNameOrGID string, fsRoot string) (*uint32, error) {
	var gid *uint32
	if groupNameOrGID != "" {
		if parsed, err := strconv.ParseUint(groupNameOrGID, 10, 32); errors.Is(err, strconv.ErrSyntax) {
			result, err := getGIDFromGroupName(groupNameOrGID, filepath.Join(fsRoot, "etc", "group"))
			if err != nil {
				return nil, fmt.Errorf("error getting GID from group name: %w", err)
			}
			gid = &result
		} else if err != nil {
			return nil, fmt.Errorf("error parsing group ID: %w", err)
		} else {
			result := uint32(parsed)
			gid = &result
		}
	}
	return gid, nil
}

func getUIDFromUserName(user string, etcPasswdPath string) (uint32, error) {
	passwd, err := os.Open(etcPasswdPath)
	if err != nil {
		return 0, fmt.Errorf("error opening passwd file: %w", err)
	}
	defer passwd.Close()

	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		userParams := strings.Split(scanner.Text(), ":")
		if len(userParams) < 3 {
			continue
		}

		if userParams[0] == user {
			uid, err := strconv.ParseUint(userParams[2], 10, 32)
			if err != nil {
				return 0, fmt.Errorf("unexpected UID in passwd file, can't parse to uint: %w", err)
			}
			return uint32(uid), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error scanning passwd file: %w", err)
	}

	return 0, fmt.Errorf("could not find UID for user %q in passwd file %q", user, etcPasswdPath)
}

func getGIDFromGroupName(group string, etcGroupPath string) (uint32, error) {
	etcGroup, err := os.Open(etcGroupPath)
	if err != nil {
		return 0, fmt.Errorf("error opening group file: %w", err)
	}
	defer etcGroup.Close()

	scanner := bufio.NewScanner(etcGroup)
	for scanner.Scan() {
		groupParams := strings.Split(scanner.Text(), ":")
		if len(groupParams) < 3 {
			continue
		}

		if groupParams[0] == group {
			gid, err := strconv.ParseUint(groupParams[2], 10, 32)
			if err != nil {
				return 0, fmt.Errorf("unexpected GID in group file, can't parse to uint: %w", err)
			}
			return uint32(gid), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error scanning group file: %w", err)
	}

	return 0, fmt.Errorf("could not find GID for group %q in passwd file %q", group, etcGroupPath)
}
