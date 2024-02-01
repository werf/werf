package container_backend

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/opencontainers/runtime-spec/specs-go"

	copyrec "github.com/werf/copy-recurse"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

type BuildahBackend struct {
	buildah buildah.Buildah
	BuildahBackendOptions
}

type BuildahBackendOptions struct {
	TmpDir string
}

func NewBuildahBackend(buildah buildah.Buildah, opts BuildahBackendOptions) *BuildahBackend {
	return &BuildahBackend{buildah: buildah, BuildahBackendOptions: opts}
}

func (backend *BuildahBackend) HasStapelBuildSupport() bool {
	return true
}

func (backend *BuildahBackend) GetDefaultPlatform() string {
	return backend.buildah.GetDefaultPlatform()
}

func (backend *BuildahBackend) GetRuntimePlatform() string {
	return backend.buildah.GetRuntimePlatform()
}

func (backend *BuildahBackend) getBuildahCommonOpts(ctx context.Context, suppressLog bool, logWriterOverride io.Writer, targetPlatform string) (opts buildah.CommonOpts) {
	if !suppressLog {
		if logWriterOverride != nil {
			opts.LogWriter = logWriterOverride
		} else {
			opts.LogWriter = logboek.Context(ctx).OutStream()
		}
	}
	opts.TargetPlatform = targetPlatform

	return
}

type containerDesc struct {
	ImageName string
	Name      string
	RootMount string
}

func (backend *BuildahBackend) createContainers(ctx context.Context, images []string, opts CommonOpts) ([]*containerDesc, error) {
	var res []*containerDesc

	for _, img := range images {
		containerID := fmt.Sprintf("werf-%s", uuid.New().String())

		if img == "" {
			panic("cannot start container for an empty image param")
		}

		_, err := backend.buildah.FromCommand(ctx, containerID, img, buildah.FromCommandOpts(backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform)))
		if err != nil {
			return nil, fmt.Errorf("unable to create container using base image %q: %w", img, err)
		}

		logboek.Context(ctx).Debug().LogF("Started container %q for image %q\n", containerID, img)
		res = append(res, &containerDesc{ImageName: img, Name: containerID})
	}

	return res, nil
}

func (backend *BuildahBackend) removeContainers(ctx context.Context, containers []*containerDesc, opts CommonOpts) error {
	for _, cont := range containers {
		if err := backend.buildah.Rm(ctx, cont.Name, buildah.RmOpts(backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform))); err != nil {
			return fmt.Errorf("unable to remove container %q: %w", cont.Name, err)
		}
	}

	return nil
}

func (backend *BuildahBackend) mountContainers(ctx context.Context, containers []*containerDesc, opts CommonOpts) error {
	for _, cont := range containers {
		containerRoot, err := backend.buildah.Mount(ctx, cont.Name, buildah.MountOpts(backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform)))
		if err != nil {
			return fmt.Errorf("unable to mount container %q root dir: %w", cont.Name, err)
		}
		cont.RootMount = containerRoot
	}

	return nil
}

func (backend *BuildahBackend) unmountContainers(ctx context.Context, containers []*containerDesc, opts CommonOpts) error {
	for _, cont := range containers {
		if err := backend.buildah.Umount(ctx, cont.Name, buildah.UmountOpts(backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform))); err != nil {
			return fmt.Errorf("container %q: %w", cont.Name, err)
		}
	}

	return nil
}

func makeScript(commands []string) []byte {
	var scriptCommands []string
	for _, c := range commands {
		scriptCommands = append(scriptCommands, fmt.Sprintf(`printf "$ %%s\n" %q`, c))
		scriptCommands = append(scriptCommands, c)
	}

	return []byte(fmt.Sprintf(`#!/bin/sh

set -e

if [ "x$_IS_REEXEC" = "x" ]; then
	if type bash >/dev/null 2>&1 ; then
		echo "# Using bash to execute commands"
		echo
		export _IS_REEXEC="1"
		exec bash $0
	else
		echo "# Using /bin/sh to execute commands"
		echo
	fi
fi

%s
`, strings.Join(scriptCommands, "\n")))
}

func (backend *BuildahBackend) applyCommands(ctx context.Context, container *containerDesc, buildVolumes, commands []string, opts CommonOpts) error {
	hostScriptPath := filepath.Join(backend.TmpDir, fmt.Sprintf("script-%s.sh", uuid.New().String()))
	if err := os.WriteFile(hostScriptPath, makeScript(commands), os.FileMode(0o555)); err != nil {
		return fmt.Errorf("unable to write script file %q: %w", hostScriptPath, err)
	}
	defer os.RemoveAll(hostScriptPath)

	logboek.Context(ctx).Default().LogF("Executing script %s\n", hostScriptPath)

	destScriptPath := "/.werf/script.sh"

	var mounts []*specs.Mount
	mounts = append(mounts, &specs.Mount{
		Type:        "bind",
		Source:      hostScriptPath,
		Destination: destScriptPath,
	})

	if m, err := makeBuildahMounts(buildVolumes); err != nil {
		return err
	} else {
		mounts = append(mounts, m...)
	}

	if err := backend.buildah.RunCommand(ctx, container.Name, []string{"sh", destScriptPath}, buildah.RunCommandOpts{
		CommonOpts:   backend.getBuildahCommonOpts(ctx, false, nil, opts.TargetPlatform),
		User:         "0:0",
		WorkingDir:   "/",
		GlobalMounts: mounts,
	}); err != nil {
		return fmt.Errorf("unable to run commands script: %w", err)
	}

	return nil
}

func (backend *BuildahBackend) CalculateDependencyImportChecksum(ctx context.Context, dependencyImport DependencyImportSpec, opts CalculateDependencyImportChecksum) (string, error) {
	// TODO(2.0): Take into account empty dirs

	var container *containerDesc
	if c, err := backend.createContainers(ctx, []string{dependencyImport.ImageName}, CommonOpts(opts)); err != nil {
		return "", err
	} else {
		container = c[0]
	}
	defer func() {
		if err := backend.removeContainers(ctx, []*containerDesc{container}, CommonOpts(opts)); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporal dependency container %q: %s\n", container.Name, err)
		}
	}()

	logboek.Context(ctx).Debug().LogF("Mounting dependency container %s\n", container.Name)
	if err := backend.mountContainers(ctx, []*containerDesc{container}, CommonOpts(opts)); err != nil {
		return "", fmt.Errorf("unable to mount build container %s: %w", container.Name, err)
	}
	defer func() {
		logboek.Context(ctx).Debug().LogF("Unmounting build container %s\n", container.Name)
		if err := backend.unmountContainers(ctx, []*containerDesc{container}, CommonOpts(opts)); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to unmount containers: %s\n", err)
		}
	}()

	fromPath := filepath.Join(container.RootMount, dependencyImport.FromPath)

	pathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:     fromPath,
		IncludeGlobs: dependencyImport.IncludePaths,
		ExcludeGlobs: dependencyImport.ExcludePaths,
	})

	var files []string

	err := filepath.Walk(fromPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		if !pathMatcher.IsDirOrSubmodulePathMatched(path) {
			if f.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		if f.IsDir() {
			return nil
		}

		if pathMatcher.IsPathMatched(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	hash := md5.New()
	sort.Strings(files)

	for _, path := range files {
		logboek.Context(ctx).Debug().LogF("Calculating checksum of container file %s\n", path)
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("unable to open file %q: %w", path, err)
		}

		fileHash := md5.New()
		if _, err := io.Copy(fileHash, f); err != nil {
			return "", fmt.Errorf("error reading file %q: %w", path, err)
		}

		if _, err := fmt.Fprintf(hash, "%x  %s\n", fileHash.Sum(nil), filepath.Join("/", util.GetRelativeToBaseFilepath(container.RootMount, path))); err != nil {
			return "", fmt.Errorf("error calculating file %q checksum: %w", path, err)
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (backend *BuildahBackend) applyDataArchives(ctx context.Context, container *containerDesc, dataArchives []DataArchiveSpec) error {
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

		var err error
		var uid, gid *uint32
		uid, gid, err = getUIDAndGID(archive.Owner, archive.Group, container.RootMount)
		if err != nil {
			return fmt.Errorf("error getting UID/GID: %w", err)
		}

		logboek.Context(ctx).Debug().LogF("Extracting archive into container path %s\n", archive.To)

		if err := util.ExtractTar(archive.Archive, extractDestPath, util.ExtractTarOptions{UID: uid, GID: gid}); err != nil {
			return fmt.Errorf("unable to extract data archive into %s: %w", archive.To, err)
		}

		if err := archive.Archive.Close(); err != nil {
			return fmt.Errorf("error closing archive data stream: %w", err)
		}
	}

	return nil
}

func (backend *BuildahBackend) applyRemoveData(ctx context.Context, container *containerDesc, removeData []RemoveDataSpec) error {
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

func (backend *BuildahBackend) applyDependenciesImports(ctx context.Context, container *containerDesc, depImports []DependencyImportSpec, opts CommonOpts) error {
	var depImages []string
	for _, imp := range depImports {
		if util.IsStringsContainValue(depImages, imp.ImageName) {
			continue
		}

		depImages = append(depImages, imp.ImageName)
	}

	logboek.Context(ctx).Debug().LogF("Creating containers for depContainers images %v\n", depImages)
	createdDepContainers, err := backend.createContainers(ctx, depImages, opts)
	if err != nil {
		return fmt.Errorf("unable to create depContainers containers: %w", err)
	}
	defer func() {
		if err := backend.removeContainers(ctx, createdDepContainers, opts); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporal depContainers containers: %s\n", err)
		}
	}()

	// NOTE: maybe it is more optimal not to mount all dependencies at once, but mount one-by-one
	logboek.Context(ctx).Debug().LogF("Mounting depContainers containers %v\n", createdDepContainers)
	if err := backend.mountContainers(ctx, createdDepContainers, opts); err != nil {
		return fmt.Errorf("unable to mount containers: %w", err)
	}
	defer func() {
		logboek.Context(ctx).Debug().LogF("Unmounting depContainers containers %v\n", createdDepContainers)
		if err := backend.unmountContainers(ctx, createdDepContainers, opts); err != nil {
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

			copyRec, err := copyrec.New(absFrom, absTo, copyrec.Options{
				MatchDir: func(path string) (copyrec.DirAction, error) {
					relPath := util.GetRelativeToBaseFilepath(absFrom, path)

					switch {
					case pathMatcher.IsPathMatched(relPath):
						return copyrec.DirMatch, nil
					case pathMatcher.ShouldGoThrough(relPath):
						return copyrec.DirFallThrough, nil
					default:
						return copyrec.DirSkip, nil
					}
				},
				MatchFile: func(path string) (bool, error) {
					relPath := util.GetRelativeToBaseFilepath(absFrom, path)
					return pathMatcher.IsPathMatched(relPath), err
				},
				UID: uid,
				GID: gid,
			})
			if err != nil {
				return fmt.Errorf("error creating recursive copy command: %w", err)
			}

			if err := copyRec.Run(ctx); err != nil {
				return fmt.Errorf("error copying dependency import files from %q to %q: %w", absFrom, absTo, err)
			}
		}
	}

	return nil
}

func (backend *BuildahBackend) BuildDockerfileStage(ctx context.Context, baseImage string, opts BuildDockerfileStageOptions, instructions ...InstructionInterface) (string, error) {
	var container *containerDesc
	if c, err := backend.createContainers(ctx, []string{baseImage}, opts.CommonOpts); err != nil {
		return "", err
	} else {
		container = c[0]
	}

	defer func() {
		if err := backend.removeContainers(ctx, []*containerDesc{container}, opts.CommonOpts); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporary build container: %s\n", err)
		}
	}()
	// TODO: cleanup orphan build containers in werf-host-cleanup procedure

	logboek.Context(ctx).Debug().LogF("Mounting build container %s\n", container.Name)
	if err := backend.mountContainers(ctx, []*containerDesc{container}, opts.CommonOpts); err != nil {
		return "", fmt.Errorf("unable to mount build container %s: %w", container.Name, err)
	}
	defer func() {
		logboek.Context(ctx).Debug().LogF("Unmounting build container %s\n", container.Name)
		if err := backend.unmountContainers(ctx, []*containerDesc{container}, opts.CommonOpts); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to unmount containers: %s\n", err)
		}
	}()

	logboek.Context(ctx).Debug().LogF("Executing commands for build container %s: %#v\n", container.Name, instructions)

	for _, instruction := range instructions {
		if err := instruction.Apply(ctx, container.Name, backend.buildah, backend.getBuildahCommonOpts(ctx, false, nil, opts.TargetPlatform), opts.BuildContextArchive); err != nil {
			return "", fmt.Errorf("unable to apply instruction %s: %w", instruction.Name(), err)
		}
	}

	logboek.Context(ctx).Debug().LogF("Committing build container %s\n", container.Name)
	imageID, err := backend.buildah.Commit(ctx, container.Name, buildah.CommitOpts{
		CommonOpts: backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform),
	})
	if err != nil {
		return "", fmt.Errorf("error committing container %s: %w", container.Name, err)
	}

	return imageID, nil
}

func (backend *BuildahBackend) BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (string, error) {
	commonOpts := CommonOpts{TargetPlatform: opts.TargetPlatform}

	var container *containerDesc
	if c, err := backend.createContainers(ctx, []string{baseImage}, commonOpts); err != nil {
		return "", err
	} else {
		container = c[0]
	}
	defer func() {
		if err := backend.removeContainers(ctx, []*containerDesc{container}, commonOpts); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporal build container: %s\n", err)
		}
	}()
	// TODO(stapel-to-buildah): cleanup orphan build containers in werf-host-cleanup procedure

	if len(opts.DependencyImportSpecs)+len(opts.DataArchiveSpecs)+len(opts.RemoveDataSpecs) > 0 {
		logboek.Context(ctx).Debug().LogF("Mounting build container %s\n", container.Name)
		if err := backend.mountContainers(ctx, []*containerDesc{container}, commonOpts); err != nil {
			return "", fmt.Errorf("unable to mount build container %s: %w", container.Name, err)
		}
		defer func() {
			logboek.Context(ctx).Debug().LogF("Unmounting build container %s\n", container.Name)
			if err := backend.unmountContainers(ctx, []*containerDesc{container}, commonOpts); err != nil {
				logboek.Context(ctx).Error().LogF("ERROR: unable to unmount containers: %s\n", err)
			}
		}()
	}

	if len(opts.DependencyImportSpecs) > 0 {
		if err := backend.applyDependenciesImports(ctx, container, opts.DependencyImportSpecs, commonOpts); err != nil {
			return "", err
		}
	}
	if len(opts.DataArchiveSpecs) > 0 {
		if err := backend.applyDataArchives(ctx, container, opts.DataArchiveSpecs); err != nil {
			return "", err
		}
	}
	if len(opts.RemoveDataSpecs) > 0 {
		if err := backend.applyRemoveData(ctx, container, opts.RemoveDataSpecs); err != nil {
			return "", err
		}
	}
	if len(opts.Commands) > 0 {
		if err := backend.applyCommands(ctx, container, opts.BuildVolumes, opts.Commands, commonOpts); err != nil {
			return "", err
		}
	}

	healthcheck, err := newHealthConfigFromString(opts.Healthcheck)
	if err != nil {
		return "", fmt.Errorf("unable to parse healthcheck %q: %w", opts.Healthcheck, err)
	}

	logboek.Context(ctx).Debug().LogF("Setting config for build container %q\n", container.Name)
	if err := backend.buildah.Config(ctx, container.Name, buildah.ConfigOpts{
		CommonOpts:  backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform),
		Labels:      opts.Labels,
		Volumes:     opts.Volumes,
		Expose:      opts.Expose,
		Envs:        opts.Envs,
		Cmd:         opts.Cmd,
		Entrypoint:  opts.Entrypoint,
		User:        opts.User,
		Workdir:     opts.Workdir,
		Healthcheck: healthcheck,
	}); err != nil {
		return "", fmt.Errorf("unable to set container %q config: %w", container.Name, err)
	}

	// TODO(stapel-to-buildah): Save container name as builtID. There is no need to commit an image here,
	//                            because buildah allows to commit and push directly container, which would happen later.
	logboek.Context(ctx).Debug().LogF("committing container %q\n", container.Name)
	imgID, err := backend.buildah.Commit(ctx, container.Name, buildah.CommitOpts{CommonOpts: backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform)})
	if err != nil {
		return "", fmt.Errorf("unable to commit container %q: %w", container.Name, err)
	}

	return imgID, nil
}

// GetImageInfo returns nil, nil if image not found.
func (backend *BuildahBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	inspect, err := backend.buildah.Inspect(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("error getting buildah inspect of %q: %w", ref, err)
	}
	if inspect == nil {
		return nil, nil
	}

	parentID := string(inspect.Docker.Parent)
	if parentID == "" {
		if id, ok := inspect.Docker.Config.Labels[image.WerfBaseImageIDLabel]; ok { // built with werf and buildah backend
			parentID = id
		}
	}

	var repository, tag, repoDigest string
	if !strings.HasPrefix(ref, "sha256:") {
		repository, tag = image.ParseRepositoryAndTag(ref)
		list, err := backend.buildah.Images(ctx, buildah.ImagesOptions{Names: []string{ref}})
		if err != nil {
			return nil, fmt.Errorf("error getting buildah info for image %q: %w", ref, err)
		}
		if len(list) > 0 {
			repoDigest = image.ExtractRepoDigest(list[0].RepoDigests, repository)
		}
	}

	return &image.Info{
		Name:              ref,
		Repository:        repository,
		RepoDigest:        repoDigest,
		Tag:               tag,
		Labels:            inspect.Docker.Config.Labels,
		CreatedAtUnixNano: inspect.Docker.Created.UnixNano(),
		OnBuild:           inspect.Docker.Config.OnBuild,
		Env:               inspect.Docker.Config.Env,
		ID:                inspect.FromImageID,
		ParentID:          parentID,
		Size:              inspect.Docker.Size,
	}, nil
}

func (backend *BuildahBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	var logWriter io.Writer
	if logboek.Context(ctx).Info().IsAccepted() {
		logWriter = logboek.Context(ctx).OutStream()
	}

	return backend.buildah.Rmi(ctx, ref, buildah.RmiOpts{
		Force:      true,
		CommonOpts: backend.getBuildahCommonOpts(ctx, false, logWriter, opts.TargetPlatform),
	})
}

func (backend *BuildahBackend) Pull(ctx context.Context, ref string, opts PullOpts) error {
	var logWriter io.Writer
	if logboek.Context(ctx).Info().IsAccepted() {
		logWriter = logboek.Context(ctx).OutStream()
	}

	return backend.buildah.Pull(
		ctx, ref,
		buildah.PullOpts(backend.getBuildahCommonOpts(ctx, false, logWriter, opts.TargetPlatform)),
	)
}

func (backend *BuildahBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	var logWriter io.Writer
	if logboek.Context(ctx).Info().IsAccepted() {
		logWriter = logboek.Context(ctx).OutStream()
	}

	return backend.buildah.Tag(
		ctx, ref, newRef,
		buildah.TagOpts(backend.getBuildahCommonOpts(ctx, false, logWriter, opts.TargetPlatform)),
	)
}

func (backend *BuildahBackend) Push(ctx context.Context, ref string, opts PushOpts) error {
	var logWriter io.Writer
	if logboek.Context(ctx).Info().IsAccepted() {
		logWriter = logboek.Context(ctx).OutStream()
	} else {
		logWriter = io.Discard
	}

	return backend.buildah.Push(
		ctx, ref,
		buildah.PushOpts(backend.getBuildahCommonOpts(ctx, false, logWriter, opts.TargetPlatform)),
	)
}

func (backend *BuildahBackend) TagImageByName(ctx context.Context, img LegacyImageInterface) error {
	if img.BuiltID() != "" {
		if err := backend.Tag(ctx, img.BuiltID(), img.Name(), TagOpts{}); err != nil {
			return fmt.Errorf("unable to tag %q as %s: %w", img.BuiltID(), img.Name(), err)
		}
	} else {
		if err := backend.RefreshImageObject(ctx, img); err != nil {
			return err
		}
	}
	return nil
}

func (backend *BuildahBackend) BuildDockerfile(ctx context.Context, dockerfileContent []byte, opts BuildDockerfileOpts) (string, error) {
	buildArgs := make(map[string]string)
	for _, argStr := range opts.BuildArgs {
		argParts := strings.SplitN(argStr, "=", 2)
		if len(argParts) < 2 {
			return "", fmt.Errorf("invalid build argument %q given, expected string in the key=value format", argStr)
		}
		buildArgs[argParts[0]] = argParts[1]
	}

	buildContextTmpDir, err := opts.BuildContextArchive.ExtractOrGetExtractedDir(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to extract build context: %w", err)
	}

	dockerfile, err := ioutil.TempFile(backend.TmpDir, "*.Dockerfile")
	if err != nil {
		return "", fmt.Errorf("error creating temporary dockerfile: %w", err)
	}

	if _, err := dockerfile.Write(dockerfileContent); err != nil {
		return "", fmt.Errorf("error writing temporary dockerfile: %w", err)
	}
	defer func() {
		if err := os.Remove(dockerfile.Name()); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporary dockerfile %s: %s\n", dockerfile.Name(), err)
		}
	}()

	return backend.buildah.BuildFromDockerfile(ctx, dockerfile.Name(), buildah.BuildFromDockerfileOpts{
		CommonOpts: backend.getBuildahCommonOpts(ctx, false, nil, opts.TargetPlatform),
		ContextDir: buildContextTmpDir,
		BuildArgs:  buildArgs,
		Target:     opts.Target,
		Labels:     opts.Labels,
	})
}

func (backend *BuildahBackend) ShouldCleanupDockerfileImage() bool {
	return false
}

func (backend *BuildahBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	if info, err := backend.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}
	return nil
}

func (backend *BuildahBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	if err := backend.Pull(ctx, img.Name(), PullOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return fmt.Errorf("unable to pull image %s: %w", img.Name(), err)
	}

	if info, err := backend.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %w", img.Name(), err)
	} else {
		img.SetInfo(info)
	}

	return nil
}

func (backend *BuildahBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", img.Name(), newImageName)).DoError(func() error {
		if err := backend.Tag(ctx, img.Name(), newImageName, TagOpts{TargetPlatform: img.GetTargetPlatform()}); err != nil {
			return fmt.Errorf("unable to tag image %s by name %s: %w", img.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", img.Name())).DoError(func() error {
			if err := backend.Rmi(ctx, img.Name(), RmiOpts{
				CommonOpts: CommonOpts{TargetPlatform: img.GetTargetPlatform()},
			}); err != nil {
				return fmt.Errorf("unable to remove image %q: %w", img.Name(), err)
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

func (backend *BuildahBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing image tag %s", img.Name())).DoError(func() error {
		if err := backend.Rmi(ctx, img.Name(), RmiOpts{
			CommonOpts: CommonOpts{TargetPlatform: img.GetTargetPlatform()},
		}); err != nil {
			return fmt.Errorf("unable to remove image %q: %w", img.Name(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (backend *BuildahBackend) String() string {
	return "buildah-backend"
}

func (backend *BuildahBackend) RemoveHostDirs(ctx context.Context, mountDir string, dirs []string) error {
	var container *containerDesc
	if c, err := backend.createContainers(ctx, []string{"alpine"}, CommonOpts{}); err != nil {
		return fmt.Errorf("unable to create container based on alpine: %w", err)
	} else {
		container = c[0]
	}
	defer func() {
		if err := backend.removeContainers(ctx, []*containerDesc{container}, CommonOpts{}); err != nil {
			logboek.Context(ctx).Error().LogF("ERROR: unable to remove temporal container %q: %s\n", container.Name, err)
		}
	}()
	var containerDirs []string
	for _, dir := range dirs {
		containerDirs = append(containerDirs, util.ToLinuxContainerPath(dir))
	}

	return backend.buildah.RunCommand(ctx, container.Name, append([]string{"rm", "-rf"}, containerDirs...), buildah.RunCommandOpts{
		User:       "0:0",
		WorkingDir: "/",
		GlobalMounts: []*specs.Mount{
			{
				Type:        "bind",
				Source:      mountDir,
				Destination: util.ToLinuxContainerPath(mountDir),
			},
		},
	})
}

func parseVolume(volume string) (string, string, error) {
	volumeParts := strings.SplitN(volume, ":", 2)
	if len(volumeParts) != 2 {
		return "", "", fmt.Errorf("expected SOURCE:DESTINATION format")
	}
	return volumeParts[0], volumeParts[1], nil
}

func makeBuildahMounts(volumes []string) ([]*specs.Mount, error) {
	var mounts []*specs.Mount

	for _, volume := range volumes {
		from, to, err := parseVolume(volume)
		if err != nil {
			return nil, fmt.Errorf("invalid volume %q: %w", volume, err)
		}

		mounts = append(mounts, &specs.Mount{
			Type:        "bind",
			Source:      from,
			Destination: to,
		})
	}

	return mounts, nil
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
func getUID(userNameOrUID, fsRoot string) (*uint32, error) {
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
func getGID(groupNameOrGID, fsRoot string) (*uint32, error) {
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

func getUIDFromUserName(user, etcPasswdPath string) (uint32, error) {
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

func getGIDFromGroupName(group, etcGroupPath string) (uint32, error) {
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

// Can return nil pointer to BuildahHealthConfig
func newHealthConfigFromString(healthcheck string) (*thirdparty.BuildahHealthConfig, error) {
	if healthcheck == "" {
		return nil, nil
	}

	dockerfile, err := parser.Parse(bytes.NewBufferString(fmt.Sprintf("HEALTHCHECK %s", healthcheck)))
	if err != nil {
		return nil, fmt.Errorf("unable to parse healthcheck instruction: %w", err)
	}

	var healthCheckNode *parser.Node
	for _, n := range dockerfile.AST.Children {
		if strings.ToLower(n.Value) == "healthcheck" {
			healthCheckNode = n
		}
	}
	if healthCheckNode == nil {
		return nil, fmt.Errorf("no valid healthcheck instruction found, got %q", healthcheck)
	}

	cmd, err := instructions.ParseCommand(healthCheckNode)
	if err != nil {
		return nil, fmt.Errorf("cannot parse healthcheck instruction: %w", err)
	}

	healthcheckcmd := cmd.(*instructions.HealthCheckCommand)
	healthconfig := &thirdparty.BuildahHealthConfig{
		Test:        healthcheckcmd.Health.Test,
		Interval:    healthcheckcmd.Health.Interval,
		Timeout:     healthcheckcmd.Health.Timeout,
		StartPeriod: healthcheckcmd.Health.StartPeriod,
		Retries:     healthcheckcmd.Health.Retries,
	}

	return healthconfig, nil
}

func (backend *BuildahBackend) Images(ctx context.Context, opts ImagesOptions) (image.ImagesList, error) {
	imagesOpts := buildah.ImagesOptions{Filters: opts.Filters}
	return backend.buildah.Images(ctx, imagesOpts)
}

func (backend *BuildahBackend) Containers(ctx context.Context, opts ContainersOptions) (image.ContainerList, error) {
	containersOpts := buildah.ContainersOptions{Filters: opts.Filters}
	return backend.buildah.Containers(ctx, containersOpts)
}

func (backend *BuildahBackend) Rm(ctx context.Context, name string, opts RmOpts) error {
	return backend.buildah.Rm(ctx, name, buildah.RmOpts{})
}

func (backend *BuildahBackend) PostManifest(ctx context.Context, ref string, opts PostManifestOpts) error {
	containerID := uuid.New().String()
	_, err := backend.buildah.FromCommand(ctx, containerID, "", buildah.FromCommandOpts(backend.getBuildahCommonOpts(ctx, true, nil, opts.TargetPlatform)))
	if err != nil {
		return fmt.Errorf("unable to create container using scratch base image: %w", err)
	}

	if err := backend.buildah.Config(ctx, containerID, buildah.ConfigOpts{Labels: opts.Labels}); err != nil {
		return fmt.Errorf("unable to configure container %q labels: %w", containerID, err)
	}

	if _, err := backend.buildah.Commit(ctx, containerID, buildah.CommitOpts{Image: ref}); err != nil {
		return fmt.Errorf("unable to commit container %q: %w", containerID, err)
	}
	return nil
}

func (backend *BuildahBackend) ClaimTargetPlatforms(ctx context.Context, targetPlatforms []string) {}
