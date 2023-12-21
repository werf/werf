package stage

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
)

type GitMapping struct {
	Name               string
	As                 string
	Branch             string
	Tag                string
	Commit             string
	Add                string
	To                 string
	Owner              string
	Group              string
	IncludePaths       []string
	ExcludePaths       []string
	StagesDependencies map[StageName][]string

	ContainerPatchesDir  string
	ContainerArchivesDir string
	ScriptsDir           string
	ContainerScriptsDir  string

	BaseCommitByPrevBuiltImageName map[string]string

	gitRepo git_repo.GitRepo
	mutexes map[string]*sync.Mutex
	mutex   sync.Mutex
}

func NewGitMapping() *GitMapping {
	return &GitMapping{
		BaseCommitByPrevBuiltImageName: map[string]string{},

		mutexes: map[string]*sync.Mutex{},
	}
}

type ContainerFileDescriptor struct {
	FilePath          string
	ContainerFilePath string
}

func (f *ContainerFileDescriptor) Open(flag int, perm os.FileMode) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(f.FilePath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	handler, err := os.OpenFile(f.FilePath, flag, perm)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

func (gm *GitMapping) getMutex(key string) *sync.Mutex {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	m, ok := gm.mutexes[key]
	if !ok {
		m = &sync.Mutex{}
		gm.mutexes[key] = m
	}

	return m
}

func (gm *GitMapping) SetGitRepo(gitRepo git_repo.GitRepo) {
	gm.gitRepo = gitRepo
}

func (gm *GitMapping) GitRepo() git_repo.GitRepo {
	return gm.gitRepo
}

func (gm *GitMapping) makeArchiveOptions(ctx context.Context, commit string) (*git_repo.ArchiveOptions, error) {
	fileRenames, err := gm.getFileRenames(ctx, commit)
	if err != nil {
		return nil, fmt.Errorf("unable to make git archive options: %w", err)
	}

	pathScope, err := gm.getPathScope(ctx, commit)
	if err != nil {
		return nil, fmt.Errorf("unable to make git archive options: %w", err)
	}

	return &git_repo.ArchiveOptions{
		PathScope:   pathScope,
		PathMatcher: gm.getPathMatcher(),
		Commit:      commit,
		FileRenames: fileRenames,
	}, nil
}

func (gm *GitMapping) makePatchOptions(ctx context.Context, fromCommit, toCommit string, withEntireFileContext, withBinary bool) (*git_repo.PatchOptions, error) {
	fileRenames, err := gm.getFileRenames(ctx, toCommit)
	if err != nil {
		return nil, fmt.Errorf("unable to make git patch options: %w", err)
	}

	pathScope, err := gm.getPathScope(ctx, toCommit)
	if err != nil {
		return nil, fmt.Errorf("unable to make git patch options: %w", err)
	}

	return &git_repo.PatchOptions{
		PathScope:             pathScope,
		PathMatcher:           gm.getPathMatcher(),
		FromCommit:            fromCommit,
		ToCommit:              toCommit,
		FileRenames:           fileRenames,
		WithEntireFileContext: withEntireFileContext,
		WithBinary:            withBinary,
	}, nil
}

func (gm *GitMapping) getPathMatcher() path_matcher.PathMatcher {
	return path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
		BasePath:     gm.Add,
		IncludeGlobs: gm.IncludePaths,
		ExcludeGlobs: gm.ExcludePaths,
	})
}

func (gm *GitMapping) getPathScope(ctx context.Context, commit string) (string, error) {
	var pathScope string

	gitAddIsDirOrSubmodule, err := gm.GitRepo().IsCommitTreeEntryDirectory(ctx, commit, gm.Add)
	if err != nil {
		return "", fmt.Errorf("unable to determine whether ls tree entry for path %q on commit %q is directory or not: %w", gm.Add, commit, err)
	}

	if gitAddIsDirOrSubmodule {
		pathScope = gm.Add
	} else {
		pathScope = filepath.ToSlash(filepath.Dir(gm.Add))
	}

	return pathScope, nil
}

func (gm *GitMapping) getFileRenames(ctx context.Context, commit string) (map[string]string, error) {
	gitAddIsDirOrSubmodule, err := gm.GitRepo().IsCommitTreeEntryDirectory(ctx, commit, gm.Add)
	if err != nil {
		return nil, fmt.Errorf("unable to determine whether ls tree entry for path %q on commit %q is directory or not: %w", gm.Add, commit, err)
	}

	fileRenames := make(map[string]string)
	if gitAddIsDirOrSubmodule {
		return fileRenames, nil
	}

	if filepath.Base(gm.Add) != filepath.Base(gm.To) {
		fileRenames[gm.Add] = filepath.Base(gm.To)
	}

	return fileRenames, nil
}

func (gm *GitMapping) IsLocal() bool {
	return gm.GitRepo().IsLocal()
}

func (gm *GitMapping) getLatestCommit(ctx context.Context) (string, error) {
	if gm.Commit != "" {
		return gm.Commit, nil
	}

	if gm.Tag != "" {
		return gm.GitRepo().TagCommit(ctx, gm.Tag)
	}

	if gm.Branch != "" {
		return gm.GitRepo().LatestBranchCommit(ctx, gm.Branch)
	}

	commit, err := gm.GitRepo().HeadCommitHash(ctx)
	if err != nil {
		return "", err
	}

	return commit, nil
}

func (gm *GitMapping) applyPatchCommand(patchFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	commands := make([]string, 0)

	var applyPatchDirectory string

	switch archiveType {
	case git_repo.FileArchive:
		applyPatchDirectory = path.Dir(gm.To)
	case git_repo.DirectoryArchive:
		applyPatchDirectory = gm.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		stapel.InstallBinPath(),
		gm.makeCredentialsOpts(),
		applyPatchDirectory,
	))

	gitCommand := fmt.Sprintf(
		"%s %s apply --ignore-whitespace --whitespace=nowarn --directory=\"%s\" --unsafe-paths %s",
		stapel.OptionalSudoCommand(gm.Owner, gm.Group),
		stapel.GitBinPath(),
		applyPatchDirectory,
		patchFile.ContainerFilePath,
	)

	commands = append(commands, strings.TrimLeft(gitCommand, " "))

	return commands, nil
}

func (gm *GitMapping) GetLatestCommitInfo(ctx context.Context, c Conveyor) (ImageCommitInfo, error) {
	res := ImageCommitInfo{}

	if commit, err := gm.getLatestCommit(ctx); err != nil {
		return ImageCommitInfo{}, err
	} else {
		res.Commit = commit
	}

	if gm.GitRepo().IsLocal() && c.GetLocalGitRepoVirtualMergeOptions().VirtualMerge {
		res.VirtualMerge = true

		fromCommit, intoCommit, err := git_repo.GetVirtualMergeParents(ctx, gm.GitRepo(), res.Commit)
		if err != nil {
			return ImageCommitInfo{}, fmt.Errorf("unable to get virtual merge commit %q parents: %w", res.Commit, err)
		}

		res.VirtualMergeFromCommit = fromCommit
		logboek.Context(ctx).Debug().LogF("Got virtual-merge-from-commit from parents of %s => %s\n", res.Commit, res.VirtualMergeFromCommit)

		res.VirtualMergeIntoCommit = intoCommit
		logboek.Context(ctx).Debug().LogF("Got virtual-merge-into-commit from parents of %s => %s\n", res.Commit, res.VirtualMergeIntoCommit)

		if res.VirtualMergeFromCommit == "" {
			return ImageCommitInfo{}, fmt.Errorf("unable to detect virtual-merge-from-commit for virtual merge commit %q", res.Commit)
		}
		if res.VirtualMergeIntoCommit == "" {
			return ImageCommitInfo{}, fmt.Errorf("unable to detect virtual-merge-into-commit for virtual merge commit %q", res.Commit)
		}
	}

	return res, nil
}

func (gm *GitMapping) AddGitCommitToImageLabels(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, stageImage *StageImage, commitInfo ImageCommitInfo) {
	addLabels := make(map[string]string)
	addLabels[gm.ImageGitCommitLabel()] = commitInfo.Commit

	if commitInfo.VirtualMerge {
		addLabels[gm.VirtualMergeLabel()] = "true"
		addLabels[gm.VirtualMergeFromCommitLabel()] = commitInfo.VirtualMergeFromCommit
		addLabels[gm.VirtualMergeIntoCommitLabel()] = commitInfo.VirtualMergeIntoCommit
	} else {
		addLabels[gm.VirtualMergeLabel()] = "false"
	}

	if len(addLabels) > 0 {
		if c.UseLegacyStapelBuilder(cb) {
			stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(addLabels)
		} else {
			stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
		}
	}
}

func (gm *GitMapping) GetBaseCommitForPrevBuiltImage(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (string, error) {
	gm.getMutex(prevBuiltImage.Image.Name()).Lock()
	defer gm.getMutex(prevBuiltImage.Image.Name()).Unlock()

	if baseCommit, hasKey := gm.BaseCommitByPrevBuiltImageName[prevBuiltImage.Image.Name()]; hasKey {
		return baseCommit, nil
	}

	prevBuiltImageCommitInfo, err := gm.GetBuiltImageCommitInfo(prevBuiltImage.Image.GetStageDescription().Info.Labels)
	if err != nil {
		return "", fmt.Errorf("error getting prev built image %s commits info: %w", prevBuiltImage.Image.Name(), err)
	}

	var baseCommit string
	if prevBuiltImageCommitInfo.VirtualMerge {
		if latestCommit, err := gm.getLatestCommit(ctx); err != nil {
			return "", err
		} else if gm.GitRepo().IsLocal() && c.GetLocalGitRepoVirtualMergeOptions().VirtualMerge && latestCommit == prevBuiltImageCommitInfo.Commit {
			baseCommit = prevBuiltImageCommitInfo.Commit
		} else {
			if detachedMergeCommit, err := gm.GitRepo().CreateDetachedMergeCommit(ctx, prevBuiltImageCommitInfo.VirtualMergeFromCommit, prevBuiltImageCommitInfo.VirtualMergeIntoCommit); err != nil {
				return "", fmt.Errorf("unable to create detached merge commit of %s into %s: %w", prevBuiltImageCommitInfo.VirtualMergeFromCommit, prevBuiltImageCommitInfo.VirtualMergeIntoCommit, err)
			} else {
				logboek.Context(ctx).Info().LogF("Created detached merge commit %s (merge %s into %s) for repo %s\n", detachedMergeCommit, prevBuiltImageCommitInfo.VirtualMergeFromCommit, prevBuiltImageCommitInfo.VirtualMergeIntoCommit, gm.GitRepo().GetName())
				baseCommit = detachedMergeCommit
			}
		}
	} else {
		baseCommit = prevBuiltImageCommitInfo.Commit
	}

	gm.BaseCommitByPrevBuiltImageName[prevBuiltImage.Image.Name()] = baseCommit
	return baseCommit, nil
}

type ImageCommitInfo struct {
	Commit                 string
	VirtualMerge           bool
	VirtualMergeFromCommit string
	VirtualMergeIntoCommit string
}

func makeInvalidImageError(label string) error {
	return fmt.Errorf("invalid image: not found commit id by label %q", label)
}

func (gm *GitMapping) GetBuiltImageCommitInfo(builtImageLabels map[string]string) (ImageCommitInfo, error) {
	res := ImageCommitInfo{}

	res.VirtualMerge = builtImageLabels[gm.VirtualMergeLabel()] == "true"
	if res.VirtualMerge {
		if commit, hasKey := builtImageLabels[gm.VirtualMergeFromCommitLabel()]; !hasKey {
			return ImageCommitInfo{}, makeInvalidImageError(gm.VirtualMergeFromCommitLabel())
		} else {
			res.VirtualMergeFromCommit = commit
		}

		if commit, hasKey := builtImageLabels[gm.VirtualMergeIntoCommitLabel()]; !hasKey {
			return ImageCommitInfo{}, makeInvalidImageError(gm.VirtualMergeIntoCommitLabel())
		} else {
			res.VirtualMergeIntoCommit = commit
		}
	}

	if commit, hasKey := builtImageLabels[gm.ImageGitCommitLabel()]; !hasKey {
		return ImageCommitInfo{}, makeInvalidImageError(gm.ImageGitCommitLabel())
	} else {
		res.Commit = commit
	}

	return res, nil
}

func (gm *GitMapping) ImageGitCommitLabel() string {
	return fmt.Sprintf("werf-git-%s-commit", gm.GetParamshash())
}

func (gm *GitMapping) VirtualMergeLabel() string {
	return fmt.Sprintf("werf-git-%s-virtual-merge", gm.GetParamshash())
}

func (gm *GitMapping) VirtualMergeFromCommitLabel() string {
	return fmt.Sprintf("werf-git-%s-virtual-merge-from-commit", gm.GetParamshash())
}

func (gm *GitMapping) VirtualMergeIntoCommitLabel() string {
	return fmt.Sprintf("werf-git-%s-virtual-merge-into-commit", gm.GetParamshash())
}

func (gm *GitMapping) baseApplyPatchCommand(ctx context.Context, fromCommit, toCommit string, prevBuiltImage *StageImage) ([]string, error) {
	archiveType := git_repo.ArchiveType(prevBuiltImage.Image.GetStageDescription().Info.Labels[gm.getArchiveTypeLabelName()])

	patchOpts, err := gm.makePatchOptions(ctx, fromCommit, toCommit, false, false)
	if err != nil {
		return nil, err
	}

	patch, err := gm.GitRepo().GetOrCreatePatch(ctx, *patchOpts)
	if err != nil {
		return nil, err
	}

	if patch.IsEmpty() {
		return nil, nil
	}

	if patch.HasBinary() {
		pathsListFile, err := gm.preparePatchPathsListFile(patch)
		if err != nil {
			return nil, fmt.Errorf("cannot create patch paths list file: %w", err)
		}

		commands := make([]string, 0)

		commands = append(commands, fmt.Sprintf(
			"%s --arg-file=%s --null %s --force",
			stapel.XargsBinPath(),
			pathsListFile.ContainerFilePath,
			stapel.RmBinPath(),
		))

		var rmEmptyChangedDirsCommands []string
		changedRelDirsByLevel := make(map[int]map[string]bool)

	getPathsLoop:
		for _, p := range patch.GetPaths() {
			targetDir := p

			for {
				targetDir = path.Dir(targetDir)
				if targetDir == "." {
					continue getPathsLoop
				}

				partsCount := len(strings.Split(targetDir, "/"))

				paths, exist := changedRelDirsByLevel[partsCount]
				if !exist {
					paths = map[string]bool{}
				} else {
					_, exist = paths[targetDir]
					if exist {
						continue getPathsLoop
					}
				}

				paths[targetDir] = true
				changedRelDirsByLevel[partsCount] = paths
			}
		}

		var levelList []int
		for level := range changedRelDirsByLevel {
			levelList = append(levelList, level)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(levelList)))
		for _, level := range levelList {
			paths := changedRelDirsByLevel[level]
			for targetRelDir := range paths {
				rmEmptyChangedDirsCommands = append(rmEmptyChangedDirsCommands, fmt.Sprintf("if [ -d %[3]s ] && [ ! \"$(%[1]s -A %[3]s)\" ]; then %[2]s -d %[3]s; fi",
					stapel.LsBinPath(),
					stapel.RmBinPath(),
					quoteShellArg(filepath.Join(gm.To, targetRelDir)),
				))
			}
		}

		archiveOpts, err := gm.makeArchiveOptions(ctx, toCommit)
		if err != nil {
			return nil, err
		}

		archive, err := gm.GitRepo().GetOrCreateArchive(ctx, *archiveOpts)
		if err != nil {
			return nil, err
		}

		archiveFile, err := gm.prepareArchiveFile(archive)
		if err != nil {
			return nil, fmt.Errorf("cannot prepare archive file: %w", err)
		}

		archiveType, err := gm.getArchiveType(ctx, toCommit)
		if err != nil {
			return nil, err
		}

		applyArchiveCommands, err := gm.applyArchiveCommand(archiveFile, archiveType)
		if err != nil {
			return nil, err
		}
		commands = append(commands, applyArchiveCommands...)

		commands = append(commands, rmEmptyChangedDirsCommands...)

		return commands, nil
	}

	patchFile, err := gm.preparePatchFile(patch)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare patch file: %w", err)
	}

	return gm.applyPatchCommand(patchFile, archiveType)
}

func quoteShellArg(arg string) string {
	if len(arg) == 0 {
		return "''"
	}

	pattern := regexp.MustCompile(`[^\w@%+=:,./-]`)
	if pattern.MatchString(arg) {
		return "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
	}

	return arg
}

func (gm *GitMapping) applyArchiveCommand(archiveFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	var unpackArchiveDirectory string
	commands := make([]string, 0)

	switch archiveType {
	case git_repo.FileArchive:
		unpackArchiveDirectory = path.Dir(gm.To)
	case git_repo.DirectoryArchive:
		unpackArchiveDirectory = gm.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		stapel.InstallBinPath(),
		gm.makeCredentialsOpts(),
		unpackArchiveDirectory,
	))

	tarCommand := fmt.Sprintf(
		"%s %s -xf %s -C \"%s\"",
		stapel.OptionalSudoCommand(gm.Owner, gm.Group),
		stapel.TarBinPath(),
		archiveFile.ContainerFilePath,
		unpackArchiveDirectory,
	)

	commands = append(commands, strings.TrimLeft(tarCommand, " "))

	return commands, nil
}

func (gm *GitMapping) PreparePatchForImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage) error {
	fromCommit, err := gm.GetBaseCommitForPrevBuiltImage(ctx, c, prevBuiltImage)
	if err != nil {
		return fmt.Errorf("unable to get base commit from built image: %w", err)
	}
	toCommitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return fmt.Errorf("unable to get latest commit info: %w", err)
	}

	if c.UseLegacyStapelBuilder(cb) {
		commands, err := gm.baseApplyPatchCommand(ctx, fromCommit, toCommitInfo.Commit, prevBuiltImage)
		if err != nil {
			return err
		}

		if err := gm.applyScript(stageImage, commands); err != nil {
			return err
		}
	} else {
		patchOpts, err := gm.makePatchOptions(ctx, fromCommit, toCommitInfo.Commit, false, false)
		if err != nil {
			return fmt.Errorf("unable to make patch options: %w", err)
		}
		logboek.Context(ctx).Debug().LogF("Creating patch from %s to %s\n", fromCommit, toCommitInfo.Commit)
		patch, err := gm.GitRepo().GetOrCreatePatch(ctx, *patchOpts)
		if err != nil {
			return fmt.Errorf("unable to create patch: %w", err)
		}
		if !patch.IsEmpty() {
			archiveOpts, err := gm.makeArchiveOptions(ctx, toCommitInfo.Commit)
			if err != nil {
				return err
			}
			logboek.Context(ctx).Debug().LogF("Creating archive for commit %s\n", toCommitInfo.Commit)
			archive, err := gm.GitRepo().GetOrCreateArchive(ctx, *archiveOpts)
			if err != nil {
				return fmt.Errorf("unable to create git archive for commit %s with path scope %s: %w", archiveOpts.Commit, archiveOpts.PathScope, err)
			}
			var archiveType container_backend.ArchiveType
			gitArchiveType, err := gm.getArchiveType(ctx, toCommitInfo.Commit)
			if err != nil {
				return fmt.Errorf("unable to determine git archive type: %w", err)
			}
			switch gitArchiveType {
			case git_repo.FileArchive:
				archiveType = container_backend.FileArchive
			case git_repo.DirectoryArchive:
				archiveType = container_backend.DirectoryArchive
			}

			tarBuf := buffer.New(64 * 1024 * 1024)
			patchArchiveReader, patchArchiveWriter := nio.Pipe(tarBuf)
			f, err := os.Open(archive.GetFilePath())
			if err != nil {
				return fmt.Errorf("unable to open archive file %q: %w", archive.GetFilePath(), err)
			}

			var includePaths []string
			for _, path := range patch.GetPaths() {
				if util.IsStringsContainValue(patch.GetPathsToRemove(), path) {
					continue
				}
				includePaths = append(includePaths, path)
			}

			go func() {
				logboek.Context(ctx).Debug().LogF("Starting archive %q filtering process, includePaths: %v\n", archive.GetFilePath(), includePaths)
				if err := filterTarArchive(ctx, f, patchArchiveWriter, includePaths); err != nil {
					logboek.Context(ctx).Error().LogF("ERROR: %s\n", err)
					panic("tar writer close failed")
				}
			}()

			logboek.Context(ctx).Debug().LogF("Adding git patch data archive with included paths: %v\n", includePaths)
			stageImage.Builder.StapelStageBuilder().AddDataArchive(patchArchiveReader, archiveType, gm.To, container_backend.AddDataArchiveOptions{
				Owner: gm.Owner,
				Group: gm.Group,
			})

			logboek.Context(ctx).Debug().LogF("Adding git paths to remove: %v\n", patch.GetPathsToRemove())
			var pathsToRemove []string
			for _, path := range patch.GetPathsToRemove() {
				pathsToRemove = append(pathsToRemove, filepath.Join(gm.To, path))
			}
			stageImage.Builder.StapelStageBuilder().RemoveData(container_backend.RemoveExactPathWithEmptyParentDirs, pathsToRemove, []string{gm.To})
		}
	}

	gm.AddGitCommitToImageLabels(ctx, c, cb, stageImage, toCommitInfo)

	return nil
}

func filterTarArchive(ctx context.Context, in io.Reader, out io.Writer, includePaths []string) (resErr error) {
	tw := tar.NewWriter(out)

	defer func() {
		err := tw.Close()
		if resErr == nil {
			resErr = err
		}
	}()

	resErr = util.CopyTar(ctx, in, tw, util.CopyTarOptions{IncludePaths: includePaths})

	return
}

func (gm *GitMapping) PrepareArchiveForImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, stageImage *StageImage) error {
	// FIXME: legacy stapel
	// FIXME: file-archive type

	commitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return fmt.Errorf("unable to get latest commit info: %w", err)
	}

	if c.UseLegacyStapelBuilder(cb) {
		commands, err := gm.baseApplyArchiveCommand(ctx, commitInfo.Commit, stageImage)
		if err != nil {
			return err
		}

		if err := gm.applyScript(stageImage, commands); err != nil {
			return err
		}
	} else {
		archiveOpts, err := gm.makeArchiveOptions(ctx, commitInfo.Commit)
		if err != nil {
			return err
		}
		archive, err := gm.GitRepo().GetOrCreateArchive(ctx, *archiveOpts)
		if err != nil {
			return fmt.Errorf("unable to create git archive for commit %s with path scope %s: %w", archiveOpts.Commit, archiveOpts.PathScope, err)
		}
		var archiveType container_backend.ArchiveType

		gitArchiveType, err := gm.getArchiveType(ctx, commitInfo.Commit)
		if err != nil {
			return fmt.Errorf("unable to determine git archive type: %w", err)
		}

		stageImage.Builder.StapelStageBuilder().AddLabels(map[string]string{gm.getArchiveTypeLabelName(): string(gitArchiveType)})

		switch gitArchiveType {
		case git_repo.FileArchive:
			archiveType = container_backend.FileArchive
		case git_repo.DirectoryArchive:
			archiveType = container_backend.DirectoryArchive
		}

		f, err := os.Open(archive.GetFilePath())
		if err != nil {
			return fmt.Errorf("unable to open archive file %q: %w", archive.GetFilePath(), err)
		}

		stageImage.Builder.StapelStageBuilder().AddDataArchive(f, archiveType, gm.To, container_backend.AddDataArchiveOptions{
			Owner: gm.Owner,
			Group: gm.Group,
		})
	}

	gm.AddGitCommitToImageLabels(ctx, c, cb, stageImage, commitInfo)

	return nil
}

func (gm *GitMapping) applyScript(stageImage *StageImage, commands []string) error {
	stageHostTmpScriptFilePath := filepath.Join(gm.ScriptsDir, gm.GetParamshash())
	containerTmpScriptFilePath := path.Join(gm.ContainerScriptsDir, gm.GetParamshash())

	if err := stapel.CreateScript(stageHostTmpScriptFilePath, commands); err != nil {
		return err
	}

	stageImage.Builder.LegacyStapelStageBuilder().Container().AddServiceRunCommands(containerTmpScriptFilePath)

	return nil
}

func (gm *GitMapping) baseApplyArchiveCommand(ctx context.Context, commit string, stageImage *StageImage) ([]string, error) {
	archiveOpts, err := gm.makeArchiveOptions(ctx, commit)
	if err != nil {
		return nil, err
	}

	archive, err := gm.GitRepo().GetOrCreateArchive(ctx, *archiveOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to create git archive for commit %s with path scope %s: %w", archiveOpts.Commit, archiveOpts.PathScope, err)
	}

	archiveFile, err := gm.prepareArchiveFile(archive)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare archive file: %w", err)
	}

	archiveType, err := gm.getArchiveType(ctx, commit)
	if err != nil {
		return nil, err
	}

	commands, err := gm.applyArchiveCommand(archiveFile, archiveType)
	if err != nil {
		return nil, err
	}

	stageImage.Builder.LegacyStapelStageBuilder().Container().ServiceCommitChangeOptions().AddLabel(map[string]string{gm.getArchiveTypeLabelName(): string(archiveType)})

	return commands, err
}

func (gm *GitMapping) StageDependenciesChecksum(ctx context.Context, c Conveyor, stageName StageName) (string, error) {
	depsPaths := gm.StagesDependencies[stageName]
	if len(depsPaths) == 0 {
		return "", nil
	}

	commitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit info: %w", err)
	}

	hash := sha256.New()
	gitMappingPathMatcher := gm.getPathMatcher()
	for _, p := range depsPaths {
		pPathMatcher := path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{
			BasePath:     gm.Add,
			IncludeGlobs: []string{p},
		})

		multiPathMatcher := path_matcher.NewMultiPathMatcher(
			gitMappingPathMatcher,
			pPathMatcher,
		)

		checksumOptions := git_repo.ChecksumOptions{
			LsTreeOptions: git_repo.LsTreeOptions{
				PathScope:   gm.Add,
				PathMatcher: multiPathMatcher,
				AllFiles:    false,
			},
			Commit: commitInfo.Commit,
		}

		checksum, err := gm.GitRepo().GetOrCreateChecksum(ctx, checksumOptions)
		if err != nil {
			return "", err
		}

		if checksum == "" {
			absDepPath := path.Join(gm.Add, p)
			logboek.Context(ctx).Warn().LogF(
				"WARNING: stage %s dependency path %q has not been found in %s git\n",
				stageName, absDepPath, gm.GitRepo().GetName(),
			)
		} else {
			hash.Write([]byte(checksum))
		}
	}
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	return checksum, nil
}

func (gm *GitMapping) PatchSize(ctx context.Context, c Conveyor, fromCommit string) (int64, error) {
	toCommitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return 0, fmt.Errorf("unable to get latest commit info: %w", err)
	}

	if fromCommit == toCommitInfo.Commit {
		return 0, nil
	}

	patchOpts, err := gm.makePatchOptions(ctx, fromCommit, toCommitInfo.Commit, true, true)
	if err != nil {
		return 0, err
	}

	patch, err := gm.GitRepo().GetOrCreatePatch(ctx, *patchOpts)
	if err != nil {
		return 0, err
	}

	fileInfo, err := os.Stat(patch.GetFilePath())
	if err != nil {
		return 0, fmt.Errorf("unable to stat temporary patch file `%s`: %w", patch.GetFilePath(), err)
	}

	return fileInfo.Size(), nil
}

func (gm *GitMapping) GetFullName() string {
	if gm.Name != "" {
		return fmt.Sprintf("%s_%s", gm.GitRepo().GetName(), gm.Name)
	}
	return gm.GitRepo().GetName()
}

func (gm *GitMapping) GetParamshash() string {
	var err error

	hash := sha256.New()

	var parts []string
	parts = append(parts, gm.GetFullName())
	parts = append(parts, ":::")
	parts = append(parts, gm.To)
	parts = append(parts, ":::")
	parts = append(parts, gm.Add)
	parts = append(parts, ":::")
	parts = append(parts, gm.IncludePaths...)
	parts = append(parts, ":::")
	parts = append(parts, gm.ExcludePaths...)
	parts = append(parts, ":::")
	parts = append(parts, gm.Owner)
	parts = append(parts, ":::")
	parts = append(parts, gm.Group)
	parts = append(parts, ":::")
	parts = append(parts, gm.Branch)
	parts = append(parts, ":::")
	parts = append(parts, gm.Tag)
	parts = append(parts, ":::")
	parts = append(parts, gm.Commit)

	for _, part := range parts {
		_, err = hash.Write([]byte(part))
		if err != nil {
			panic(fmt.Sprintf("error calculating sha256 of `%s`: %s", part, err))
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (gm *GitMapping) GetPatchContent(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (string, error) {
	fromCommit, err := gm.GetBaseCommitForPrevBuiltImage(ctx, c, prevBuiltImage)
	if err != nil {
		return "", fmt.Errorf("unable to get base commit from built image for git mapping %s: %w", gm.GetFullName(), err)
	}

	toCommitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit info: %w", err)
	}

	if fromCommit == toCommitInfo.Commit {
		return "", nil
	}

	patchOpts, err := gm.makePatchOptions(ctx, fromCommit, toCommitInfo.Commit, false, false)
	if err != nil {
		return "", err
	}

	patch, err := gm.GitRepo().GetOrCreatePatch(ctx, *patchOpts)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(patch.GetFilePath())
	if err != nil {
		return "", fmt.Errorf("error reading patch file %s: %w", patch.GetFilePath(), err)
	}
	return string(data), nil
}

func (gm *GitMapping) IsPatchEmpty(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (bool, error) {
	fromCommit, err := gm.GetBaseCommitForPrevBuiltImage(ctx, c, prevBuiltImage)
	if err != nil {
		return false, fmt.Errorf("unable to get base commit from built image for git mapping %s: %w", gm.GetFullName(), err)
	}

	toCommitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return false, fmt.Errorf("unable to get latest commit info: %w", err)
	}

	return gm.baseIsPatchEmpty(ctx, fromCommit, toCommitInfo.Commit)
}

func (gm *GitMapping) baseIsPatchEmpty(ctx context.Context, fromCommit, toCommit string) (bool, error) {
	if fromCommit == toCommit {
		return true, nil
	}

	patchOpts, err := gm.makePatchOptions(ctx, fromCommit, toCommit, false, false)
	if err != nil {
		return false, err
	}

	patch, err := gm.GitRepo().GetOrCreatePatch(ctx, *patchOpts)
	if err != nil {
		return false, err
	}

	return patch.IsEmpty(), nil
}

func (gm *GitMapping) prepareArchiveFile(archive git_repo.Archive) (*ContainerFileDescriptor, error) {
	return &ContainerFileDescriptor{
		FilePath:          archive.GetFilePath(),
		ContainerFilePath: path.Join(gm.ContainerArchivesDir, filepath.ToSlash(util.GetRelativeToBaseFilepath(git_repo.CommonGitDataManager.GetArchivesCacheDir(), archive.GetFilePath()))),
	}, nil
}

func (gm *GitMapping) preparePatchPathsListFile(patch git_repo.Patch) (*ContainerFileDescriptor, error) {
	pathsListFilePath := filepath.Join(filepath.Dir(patch.GetFilePath()), fmt.Sprintf("%s.%s.paths_list", filepath.Base(patch.GetFilePath()), gm.GetParamshash()))
	containerFilePath := path.Join(gm.ContainerPatchesDir, filepath.ToSlash(util.GetRelativeToBaseFilepath(git_repo.CommonGitDataManager.GetPatchesCacheDir(), pathsListFilePath)))

	fileDesc := &ContainerFileDescriptor{
		FilePath:          pathsListFilePath,
		ContainerFilePath: containerFilePath,
	}

	fileExists := true
	if _, err := os.Stat(fileDesc.FilePath); os.IsNotExist(err) {
		fileExists = false
	} else if err != nil {
		return nil, fmt.Errorf("unable to get stat of path %s: %w", fileDesc.FilePath, err)
	}

	if fileExists {
		return fileDesc, nil
	}

	f, err := fileDesc.Open(os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return nil, fmt.Errorf("unable to open file `%s`: %w", fileDesc.FilePath, err)
	}

	fullPaths := make([]string, 0)
	for _, p := range patch.GetPaths() {
		fullPaths = append(fullPaths, path.Join(gm.To, p))
	}

	pathsData := strings.Join(fullPaths, "\000")
	_, err = f.Write([]byte(pathsData))
	if err != nil {
		return nil, fmt.Errorf("unable to write file `%s`: %w", fileDesc.FilePath, err)
	}

	err = f.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to close file `%s`: %w", fileDesc.FilePath, err)
	}

	return fileDesc, nil
}

func (gm *GitMapping) preparePatchFile(patch git_repo.Patch) (*ContainerFileDescriptor, error) {
	return &ContainerFileDescriptor{
		FilePath:          patch.GetFilePath(),
		ContainerFilePath: path.Join(gm.ContainerPatchesDir, filepath.ToSlash(util.GetRelativeToBaseFilepath(git_repo.CommonGitDataManager.GetPatchesCacheDir(), patch.GetFilePath()))),
	}, nil
}

func (gm *GitMapping) makeCredentialsOpts() string {
	opts := make([]string, 0)

	if gm.Owner != "" {
		opts = append(opts, fmt.Sprintf("--owner=%s", gm.Owner))
	}
	if gm.Group != "" {
		opts = append(opts, fmt.Sprintf("--group=%s", gm.Group))
	}

	return strings.Join(opts, " ")
}

func (gm *GitMapping) getArchiveTypeLabelName() string {
	return fmt.Sprintf("werf-git-%s-type", gm.GetParamshash())
}

func (gm *GitMapping) getArchiveType(ctx context.Context, commit string) (git_repo.ArchiveType, error) {
	archiveTypeIsDirectory, err := gm.GitRepo().IsCommitTreeEntryDirectory(ctx, commit, gm.Add)
	if err != nil {
		return "", fmt.Errorf("unable to determine git mapping archive type for commit %q: %w", commit, err)
	}

	if archiveTypeIsDirectory {
		return git_repo.DirectoryArchive, nil
	}

	return git_repo.FileArchive, nil
}

func (gm *GitMapping) isEmpty(ctx context.Context, c Conveyor) (bool, error) {
	commitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return true, fmt.Errorf("unable to get latest commit info: %w", err)
	}

	pathScope, err := gm.getPathScope(ctx, commitInfo.Commit)
	if err != nil {
		return true, fmt.Errorf("unable to get path scope: %w", err)
	}

	isAnyMatchesByGitAdd, err := gm.GitRepo().IsAnyCommitTreeEntriesMatched(ctx, commitInfo.Commit, pathScope, gm.getPathMatcher(), true)
	if err != nil {
		return true, err
	}

	return !isAnyMatchesByGitAdd, nil
}
