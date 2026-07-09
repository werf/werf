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
	"sync"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/path_matcher"
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
		Owner:       gm.Owner,
		Group:       gm.Group,
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

func (gm *GitMapping) getCommit(ctx context.Context) (string, error) {
	if gm.Commit != "" {
		exist, err := gm.GitRepo().IsCommitExists(ctx, gm.Commit)
		if err != nil {
			return "", fmt.Errorf("unable to check commit %q existence in repository %s: %w", gm.Commit, gm.GitRepo().GetName(), err)
		}
		if !exist {
			return "", fmt.Errorf("commit %q not found in repository %s. Ensure it has not been squashed or rebased", gm.Commit, gm.GitRepo().GetName())
		}

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

func (gm *GitMapping) GetLatestCommitInfo(ctx context.Context, c Conveyor) (ImageCommitInfo, error) {
	res := ImageCommitInfo{}

	if commit, err := gm.getCommit(ctx); err != nil {
		return ImageCommitInfo{}, err
	} else {
		res.Commit = commit
	}

	return res, nil
}

func (gm *GitMapping) AddGitCommitToImageLabels(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, stageImage *StageImage, commitInfo ImageCommitInfo) {
	addLabels := map[string]string{
		gm.ImageGitCommitLabel(): commitInfo.Commit,
	}

	if len(addLabels) > 0 {
		stageImage.Builder.StapelStageBuilder().AddLabels(addLabels)
	}
}

func (gm *GitMapping) GetBaseCommitForPrevBuiltImage(ctx context.Context, c Conveyor, prevBuiltImage *StageImage) (string, error) {
	gm.getMutex(prevBuiltImage.Image.Name()).Lock()
	defer gm.getMutex(prevBuiltImage.Image.Name()).Unlock()

	if baseCommit, hasKey := gm.BaseCommitByPrevBuiltImageName[prevBuiltImage.Image.Name()]; hasKey {
		return baseCommit, nil
	}

	prevBuiltImageCommitInfo, err := gm.GetBuiltImageCommitInfo(prevBuiltImage.Image.GetStageDesc().Info.Labels)
	if err != nil {
		return "", fmt.Errorf("error getting prev built image %s commits info: %w", prevBuiltImage.Image.Name(), err)
	}

	baseCommit := prevBuiltImageCommitInfo.Commit

	gm.BaseCommitByPrevBuiltImageName[prevBuiltImage.Image.Name()] = baseCommit
	return baseCommit, nil
}

type ImageCommitInfo struct {
	Commit string
}

func makeInvalidImageError(label string) error {
	return fmt.Errorf("invalid image: not found commit id by label %q", label)
}

func (gm *GitMapping) GetBuiltImageCommitInfo(builtImageLabels map[string]string) (ImageCommitInfo, error) {
	commit, hasKey := builtImageLabels[gm.ImageGitCommitLabel()]
	if !hasKey {
		return ImageCommitInfo{}, makeInvalidImageError(gm.ImageGitCommitLabel())
	}

	return ImageCommitInfo{Commit: commit}, nil
}

func (gm *GitMapping) ImageGitCommitLabel() string {
	return fmt.Sprintf("werf-git-%s-commit", gm.GetParamshash())
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
	commitInfo, err := gm.GetLatestCommitInfo(ctx, c)
	if err != nil {
		return fmt.Errorf("unable to get latest commit info: %w", err)
	}

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

	gm.AddGitCommitToImageLabels(ctx, c, cb, stageImage, commitInfo)

	return nil
}

func (gm *GitMapping) StageDependenciesChecksum(ctx context.Context, c Conveyor, stageName StageName) (string, error) {
	depsPaths := gm.StagesDependencies[stageName]
	if len(depsPaths) == 0 {
		depsPaths = []string{"**/*"}
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
