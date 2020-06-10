package stage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/util"
)

type GitRepoCache struct {
	Patches   map[string]git_repo.Patch
	Checksums map[string]git_repo.Checksum
	Archives  map[string]git_repo.Archive
}

func objectToHashKey(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("unable to marshal object %#v: %s", obj, err))
	}
	return util.Sha256Hash(string(data))
}

func (cache *GitRepoCache) Terminate() error {
	for _, patch := range cache.Patches {
		_ = os.RemoveAll(patch.GetFilePath())
	}
	for _, archive := range cache.Archives {
		_ = os.RemoveAll(archive.GetFilePath())
	}
	return nil
}

type GitMapping struct {
	GitRepoInterface git_repo.GitRepo
	LocalGitRepo     *git_repo.Local
	RemoteGitRepo    *git_repo.Remote
	GitRepoCache     *GitRepoCache

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

	PatchesDir           string
	ContainerPatchesDir  string
	ArchivesDir          string
	ContainerArchivesDir string
	ScriptsDir           string
	ContainerScriptsDir  string

	BaseCommitByPrevBuiltImageName map[string]string
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

func (gm *GitMapping) GitRepo() git_repo.GitRepo {
	if gm.GitRepoInterface != nil {
		return gm.GitRepoInterface
	} else if gm.LocalGitRepo != nil {
		return gm.LocalGitRepo
	} else if gm.RemoteGitRepo != nil {
		return gm.RemoteGitRepo
	}

	panic("GitRepo not initialized")
}

func (gm *GitMapping) getOrCreateChecksum(opts git_repo.ChecksumOptions) (git_repo.Checksum, error) {
	if _, hasKey := gm.GitRepoCache.Checksums[objectToHashKey(opts)]; !hasKey {
		checksum, err := gm.GitRepo().Checksum(opts)
		if err != nil {
			return nil, err
		}
		gm.GitRepoCache.Checksums[objectToHashKey(opts)] = checksum
	}
	return gm.GitRepoCache.Checksums[objectToHashKey(opts)], nil
}

func (gm *GitMapping) getOrCreateArchive(opts git_repo.ArchiveOptions) (git_repo.Archive, error) {
	if _, hasKey := gm.GitRepoCache.Archives[objectToHashKey(opts)]; !hasKey {
		archive, err := gm.createArchive(opts)
		if err != nil {
			return nil, err
		}
		gm.GitRepoCache.Archives[objectToHashKey(opts)] = archive
	}
	return gm.GitRepoCache.Archives[objectToHashKey(opts)], nil
}

func (gm *GitMapping) createArchive(opts git_repo.ArchiveOptions) (git_repo.Archive, error) {
	var res git_repo.Archive

	err := logboek.Info.LogProcess(fmt.Sprintf("Creating archive for commit %s of %s git mapping %s", opts.Commit, gm.GitRepo().GetName(), gm.Add), logboek.LevelLogProcessOptions{}, func() error {
		archive, err := gm.GitRepo().CreateArchive(opts)
		if err != nil {
			return err
		}

		res = archive

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (gm *GitMapping) getOrCreatePatch(opts git_repo.PatchOptions) (git_repo.Patch, error) {
	if _, hasKey := gm.GitRepoCache.Patches[objectToHashKey(opts)]; !hasKey {
		patch, err := gm.createPatch(opts)
		if err != nil {
			return nil, err
		}
		gm.GitRepoCache.Patches[objectToHashKey(opts)] = patch
	}
	return gm.GitRepoCache.Patches[objectToHashKey(opts)], nil
}

func (gm *GitMapping) createPatch(opts git_repo.PatchOptions) (git_repo.Patch, error) {
	var res git_repo.Patch

	logProcessMsg := fmt.Sprintf("Creating patch %s..%s for %s git mapping %s", opts.FromCommit, opts.ToCommit, gm.GitRepo().GetName(), gm.Add)
	err := logboek.Info.LogProcess(logProcessMsg, logboek.LevelLogProcessOptions{}, func() error {
		patch, err := gm.GitRepo().CreatePatch(opts)
		if err != nil {
			return err
		}

		res = patch

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (gm *GitMapping) getRepoFilterOptions() git_repo.FilterOptions {
	return git_repo.FilterOptions{
		BasePath:     gm.Add,
		IncludePaths: gm.IncludePaths,
		ExcludePaths: gm.ExcludePaths,
	}
}

func (gm *GitMapping) IsLocal() bool {
	if gm.LocalGitRepo != nil {
		return true
	} else {
		return false
	}
}

func (gm *GitMapping) getLatestCommit() (string, error) {
	if gm.Commit != "" {
		return gm.Commit, nil
	}

	if gm.Tag != "" {
		return gm.GitRepo().TagCommit(gm.Tag)
	}

	if gm.Branch != "" {
		return gm.GitRepo().LatestBranchCommit(gm.Branch)
	}

	commit, err := gm.GitRepo().HeadCommit()
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
		"%s %s apply --whitespace=nowarn --directory=\"%s\" --unsafe-paths %s",
		stapel.OptionalSudoCommand(gm.Owner, gm.Group),
		stapel.GitBinPath(),
		applyPatchDirectory,
		patchFile.ContainerFilePath,
	)

	commands = append(commands, strings.TrimLeft(gitCommand, " "))

	return commands, nil
}

func (gm *GitMapping) ApplyPatchCommand(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	fromCommit, err := gm.GetBaseCommitForPrevBuiltImage(c, prevBuiltImage)
	if err != nil {
		return fmt.Errorf("unable to get base commit from built image: %s", err)
	}

	toCommitInfo, err := gm.GetLatestCommitInfo(c)
	if err != nil {
		return fmt.Errorf("unable to get latest commit info: %s", err)
	}

	commands, err := gm.baseApplyPatchCommand(fromCommit, toCommitInfo.Commit, prevBuiltImage)
	if err != nil {
		return err
	}

	if err := gm.applyScript(image, commands); err != nil {
		return err
	}

	gm.AddGitCommitToImageLabels(image, toCommitInfo)

	return nil
}

func (gm *GitMapping) GetLatestCommitInfo(c Conveyor) (ImageCommitInfo, error) {
	res := ImageCommitInfo{}

	if commit, err := gm.getLatestCommit(); err != nil {
		return ImageCommitInfo{}, err
	} else {
		res.Commit = commit
	}

	if _, isLocal := gm.GitRepo().(*git_repo.Local); isLocal && c.GetLocalGitRepoVirtualMergeOptions().VirtualMerge {
		res.VirtualMerge = true
		res.VirtualMergeFromCommit = c.GetLocalGitRepoVirtualMergeOptions().VirtualMergeFromCommit
		res.VirtualMergeIntoCommit = c.GetLocalGitRepoVirtualMergeOptions().VirtualMergeIntoCommit

		if res.VirtualMergeFromCommit == "" || res.VirtualMergeIntoCommit == "" {
			if parents, err := gm.GitRepo().GetMergeCommitParents(res.Commit); err != nil {
				return ImageCommitInfo{}, fmt.Errorf("unable to get virtual merge commit %s parents for git repo %s: %s", res.Commit, gm.GitRepo().GetName(), err)
			} else if len(parents) == 2 {
				if res.VirtualMergeIntoCommit == "" {
					res.VirtualMergeIntoCommit = parents[0]
					logboek.Debug.LogF("Got virtual-merge-into-commit from parents of %s => %s\n", res.Commit, res.VirtualMergeIntoCommit)
				}
				if res.VirtualMergeFromCommit == "" {
					res.VirtualMergeFromCommit = parents[1]
					logboek.Debug.LogF("Got virtual-merge-from-commit from parents of %s => %s\n", res.Commit, res.VirtualMergeFromCommit)
				}
			}
		}

		if res.VirtualMergeFromCommit == "" {
			return ImageCommitInfo{}, fmt.Errorf("unable to detect --virtual-merge-from-commit for virtual merge commit %s", res.Commit)
		}
		if res.VirtualMergeIntoCommit == "" {
			return ImageCommitInfo{}, fmt.Errorf("unable to detect --virtual-merge-into-commit for virtual merge commit %s", res.Commit)
		}
	}

	return res, nil
}

func (gm *GitMapping) AddGitCommitToImageLabels(image container_runtime.ImageInterface, commitInfo ImageCommitInfo) {
	image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
		gm.ImageGitCommitLabel(): commitInfo.Commit,
	})

	if commitInfo.VirtualMerge {
		image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
			gm.VirtualMergeLabel():           "true",
			gm.VirtualMergeFromCommitLabel(): commitInfo.VirtualMergeFromCommit,
			gm.VirtualMergeIntoCommitLabel(): commitInfo.VirtualMergeIntoCommit,
		})
	}
}

func (gm *GitMapping) GetBaseCommitForPrevBuiltImage(c Conveyor, prevBuiltImage container_runtime.ImageInterface) (string, error) {
	if baseCommit, hasKey := gm.BaseCommitByPrevBuiltImageName[prevBuiltImage.Name()]; hasKey {
		return baseCommit, nil
	}

	prevBuiltImageCommitInfo, err := gm.GetBuiltImageCommitInfo(prevBuiltImage.GetStageDescription().Info.Labels)
	if err != nil {
		return "", fmt.Errorf("error getting prev built image %s commits info: %s", prevBuiltImage.Name(), err)
	}

	var baseCommit string
	if prevBuiltImageCommitInfo.VirtualMerge {
		if latestCommit, err := gm.getLatestCommit(); err != nil {
			return "", err
		} else if _, isLocal := gm.GitRepo().(*git_repo.Local); isLocal && c.GetLocalGitRepoVirtualMergeOptions().VirtualMerge && latestCommit == prevBuiltImageCommitInfo.Commit {
			baseCommit = prevBuiltImageCommitInfo.Commit
		} else {
			if detachedMergeCommit, err := gm.GitRepo().CreateDetachedMergeCommit(prevBuiltImageCommitInfo.VirtualMergeFromCommit, prevBuiltImageCommitInfo.VirtualMergeIntoCommit); err != nil {
				return "", fmt.Errorf("unable to create detached merge commit of %s into %s: %s", prevBuiltImageCommitInfo.VirtualMergeFromCommit, prevBuiltImageCommitInfo.VirtualMergeIntoCommit, err)
			} else {
				logboek.Info.LogF("Created detached merge commit %s (merge %s into %s) for repo %s\n", detachedMergeCommit, prevBuiltImageCommitInfo.VirtualMergeFromCommit, prevBuiltImageCommitInfo.VirtualMergeIntoCommit, gm.GitRepo().GetName())
				baseCommit = detachedMergeCommit
			}
		}
	} else {
		baseCommit = prevBuiltImageCommitInfo.Commit
	}

	gm.BaseCommitByPrevBuiltImageName[prevBuiltImage.Name()] = baseCommit
	return baseCommit, nil
}

type ImageCommitInfo struct {
	Commit string
	VirtualMergeOptions
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

func (gm *GitMapping) baseApplyPatchCommand(fromCommit, toCommit string, prevBuiltImage container_runtime.ImageInterface) ([]string, error) {
	archiveType := git_repo.ArchiveType(prevBuiltImage.GetStageDescription().Info.Labels[gm.getArchiveTypeLabelName()])

	patchOpts := git_repo.PatchOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}

	patch, err := gm.getOrCreatePatch(patchOpts)
	if err != nil {
		return nil, err
	}

	if patch.IsEmpty() {
		return nil, nil
	}

	if patch.HasBinary() {
		pathsListFile, err := gm.preparePatchPathsListFile(patchOpts, patch)
		if err != nil {
			return nil, fmt.Errorf("cannot create patch paths list file: %s", err)
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

		archiveOpts := git_repo.ArchiveOptions{
			FilterOptions: gm.getRepoFilterOptions(),
			Commit:        toCommit,
		}

		archive, err := gm.getOrCreateArchive(archiveOpts)
		if err != nil {
			return nil, err
		}

		if archive.IsEmpty() {
			commands = append(commands, rmEmptyChangedDirsCommands...)
			return commands, nil
		}

		archiveFile, err := gm.prepareArchiveFile(archiveOpts, archive)
		if err != nil {
			return nil, fmt.Errorf("cannot prepare archive file: %s", err)
		}

		archiveType := archive.GetType()

		applyArchiveCommands, err := gm.applyArchiveCommand(archiveFile, archiveType)
		if err != nil {
			return nil, err
		}
		commands = append(commands, applyArchiveCommands...)

		commands = append(commands, rmEmptyChangedDirsCommands...)

		return commands, nil
	}

	patchFile, err := gm.preparePatchFile(patchOpts, patch)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare patch file: %s", err)
	}

	return gm.applyPatchCommand(patchFile, archiveType)
}

func quoteShellArg(arg string) string {
	if len(arg) == 0 {
		return "''"
	}

	pattern := regexp.MustCompile(`[^\w@%+=:,./-]`)
	if pattern.MatchString(arg) {
		return "'" + strings.Replace(arg, "'", "'\"'\"'", -1) + "'"
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

func (gm *GitMapping) ApplyArchiveCommand(c Conveyor, image container_runtime.ImageInterface) error {
	commitInfo, err := gm.GetLatestCommitInfo(c)
	if err != nil {
		return fmt.Errorf("unable to get latest commit info: %s", err)
	}

	commands, err := gm.baseApplyArchiveCommand(commitInfo.Commit, image)
	if err != nil {
		return err
	}

	if err := gm.applyScript(image, commands); err != nil {
		return err
	}

	gm.AddGitCommitToImageLabels(image, commitInfo)

	return nil
}

func (gm *GitMapping) applyScript(image container_runtime.ImageInterface, commands []string) error {
	stageHostTmpScriptFilePath := filepath.Join(gm.ScriptsDir, gm.GetParamshash())
	containerTmpScriptFilePath := path.Join(gm.ContainerScriptsDir, gm.GetParamshash())

	if err := stapel.CreateScript(stageHostTmpScriptFilePath, commands); err != nil {
		return err
	}

	image.Container().AddServiceRunCommands(containerTmpScriptFilePath)

	return nil
}

func (gm *GitMapping) baseApplyArchiveCommand(commit string, image container_runtime.ImageInterface) ([]string, error) {
	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		Commit:        commit,
	}

	archive, err := gm.getOrCreateArchive(archiveOpts)
	if err != nil {
		return nil, err
	}

	if archive.IsEmpty() {
		return nil, nil
	}

	archiveFile, err := gm.prepareArchiveFile(archiveOpts, archive)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare archive file: %s", err)
	}

	archiveType := archive.GetType()

	commands, err := gm.applyArchiveCommand(archiveFile, archiveType)
	if err != nil {
		return nil, err
	}

	image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{gm.getArchiveTypeLabelName(): string(archiveType)})

	return commands, err
}

func (gm *GitMapping) StageDependenciesChecksum(c Conveyor, stageName StageName) (string, error) {
	depsPaths := gm.StagesDependencies[stageName]
	if len(depsPaths) == 0 {
		return "", nil
	}

	commitInfo, err := gm.GetLatestCommitInfo(c)
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit info: %s", err)
	}

	checksumOpts := git_repo.ChecksumOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		Paths:         depsPaths,
		Commit:        commitInfo.Commit,
	}

	checksum, err := gm.getOrCreateChecksum(checksumOpts)
	if err != nil {
		return "", err
	}

	for _, p := range checksum.GetNoMatchPaths() {
		logboek.LogWarnF(
			"WARNING: stage %s dependency path %s has not been found in %s git\n",
			stageName, p, gm.GitRepo().GetName(),
		)
	}

	return checksum.String(), nil
}

func (gm *GitMapping) PatchSize(c Conveyor, fromCommit string) (int64, error) {
	toCommitInfo, err := gm.GetLatestCommitInfo(c)
	if err != nil {
		return 0, fmt.Errorf("unable to get latest commit info: %s", err)
	}

	if fromCommit == toCommitInfo.Commit {
		return 0, nil
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions:         gm.getRepoFilterOptions(),
		FromCommit:            fromCommit,
		ToCommit:              toCommitInfo.Commit,
		WithEntireFileContext: true,
		WithBinary:            true,
	}

	patch, err := gm.getOrCreatePatch(patchOpts)
	if err != nil {
		return 0, err
	}

	fileInfo, err := os.Stat(patch.GetFilePath())
	if err != nil {
		return 0, fmt.Errorf("unable to stat temporary patch file `%s`: %s", patch.GetFilePath(), err)
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
	parts = append(parts, formatParamshashPaths(gm.Add)...)
	parts = append(parts, ":::")
	parts = append(parts, formatParamshashPaths(gm.IncludePaths...)...)
	parts = append(parts, ":::")
	parts = append(parts, formatParamshashPaths(gm.ExcludePaths...)...)
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

func formatParamshashPaths(paths ...string) []string {
	var resultPaths []string
	for _, p := range paths {
		resultPaths = append(resultPaths, filepath.ToSlash(p))
	}

	return resultPaths
}

func (gm *GitMapping) GetPatchContent(c Conveyor, prevBuiltImage container_runtime.ImageInterface) (string, error) {
	fromCommit, err := gm.GetBaseCommitForPrevBuiltImage(c, prevBuiltImage)
	if err != nil {
		return "", fmt.Errorf("unable to get base commit from built image for git mapping %s: %s", gm.GetFullName(), err)
	}

	toCommitInfo, err := gm.GetLatestCommitInfo(c)
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit info: %s", err)
	}

	if fromCommit == toCommitInfo.Commit {
		return "", nil
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommitInfo.Commit,
	}
	patch, err := gm.getOrCreatePatch(patchOpts)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(patch.GetFilePath())
	if err != nil {
		return "", fmt.Errorf("error reading patch file %s: %s", patch.GetFilePath(), err)
	}
	return string(data), nil
}

func (gm *GitMapping) IsPatchEmpty(c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
	fromCommit, err := gm.GetBaseCommitForPrevBuiltImage(c, prevBuiltImage)
	if err != nil {
		return false, fmt.Errorf("unable to get base commit from built image for git mapping %s: %s", gm.GetFullName(), err)
	}

	toCommitInfo, err := gm.GetLatestCommitInfo(c)
	if err != nil {
		return false, fmt.Errorf("unable to get latest commit info: %s", err)
	}

	return gm.baseIsPatchEmpty(fromCommit, toCommitInfo.Commit)
}

func (gm *GitMapping) baseIsPatchEmpty(fromCommit, toCommit string) (bool, error) {
	if fromCommit == toCommit {
		return true, nil
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}

	patch, err := gm.getOrCreatePatch(patchOpts)
	if err != nil {
		return false, err
	}

	return patch.IsEmpty(), nil
}

func (gm *GitMapping) IsEmpty(c Conveyor) (bool, error) {
	commitInfo, err := gm.GetLatestCommitInfo(c)
	if err != nil {
		return false, fmt.Errorf("unable to get latest commit info: %s", err)
	}

	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		Commit:        commitInfo.Commit,
	}

	archive, err := gm.getOrCreateArchive(archiveOpts)
	if err != nil {
		return false, err
	}

	return archive.IsEmpty(), nil
}

func (gm *GitMapping) getArchiveFileDescriptor(archiveOpts git_repo.ArchiveOptions) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s.tar", objectToHashKey(archiveOpts))

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gm.ArchivesDir, fileName),
		ContainerFilePath: path.Join(gm.ContainerArchivesDir, fileName),
	}
}

func (gm *GitMapping) prepareArchiveFile(archiveOpts git_repo.ArchiveOptions, archive git_repo.Archive) (*ContainerFileDescriptor, error) {
	fileDesc := gm.getArchiveFileDescriptor(archiveOpts)

	fileExists := true
	if _, err := os.Stat(fileDesc.FilePath); os.IsNotExist(err) {
		fileExists = false
	} else if err != nil {
		return nil, fmt.Errorf("unable to get stat of path %s: %s", fileDesc.FilePath, err)
	}

	if fileExists {
		return fileDesc, nil
	}

	if err := os.MkdirAll(filepath.Dir(fileDesc.FilePath), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %s: %s", filepath.Dir(fileDesc.FilePath), err)
	}

	if err := os.Link(archive.GetFilePath(), fileDesc.FilePath); err != nil {
		return nil, fmt.Errorf("unable to create hardlink %s to %s: %s", fileDesc.FilePath, archive.GetFilePath(), err)
	}

	return fileDesc, nil
}

func (gm *GitMapping) preparePatchPathsListFile(patchOpts git_repo.PatchOptions, patch git_repo.Patch) (*ContainerFileDescriptor, error) {
	fileDesc := gm.getPatchPathsListFileDescriptor(patchOpts)

	fileExists := true
	if _, err := os.Stat(fileDesc.FilePath); os.IsNotExist(err) {
		fileExists = false
	} else if err != nil {
		return nil, fmt.Errorf("unable to get stat of path %s: %s", fileDesc.FilePath, err)
	}

	if fileExists {
		return fileDesc, nil
	}

	f, err := fileDesc.Open(os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("unable to open file `%s`: %s", fileDesc.FilePath, err)
	}

	fullPaths := make([]string, 0)
	for _, p := range patch.GetPaths() {
		fullPaths = append(fullPaths, path.Join(gm.To, p))
	}

	pathsData := strings.Join(fullPaths, "\000")
	_, err = f.Write([]byte(pathsData))
	if err != nil {
		return nil, fmt.Errorf("unable to write file `%s`: %s", fileDesc.FilePath, err)
	}

	err = f.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to close file `%s`: %s", fileDesc.FilePath, err)
	}

	return fileDesc, nil
}

func (gm *GitMapping) preparePatchFile(patchOpts git_repo.PatchOptions, patch git_repo.Patch) (*ContainerFileDescriptor, error) {
	fileDesc := gm.getPatchFileDescriptor(patchOpts)

	fileExists := true
	if _, err := os.Stat(fileDesc.FilePath); os.IsNotExist(err) {
		fileExists = false
	} else if err != nil {
		return nil, fmt.Errorf("unable to get stat of path %s: %s", fileDesc.FilePath, err)
	}

	if fileExists {
		return fileDesc, nil
	}

	if err := os.MkdirAll(filepath.Dir(fileDesc.FilePath), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %s: %s", filepath.Dir(fileDesc.FilePath), err)
	}

	if err := os.Link(patch.GetFilePath(), fileDesc.FilePath); err != nil {
		return nil, fmt.Errorf("unable to create hardlink %s to %s: %s", fileDesc.FilePath, patch.GetFilePath(), err)
	}

	return fileDesc, nil
}

func (gm *GitMapping) getPatchPathsListFileDescriptor(patchOpts git_repo.PatchOptions) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s.paths_list", objectToHashKey(patchOpts))

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gm.PatchesDir, fileName),
		ContainerFilePath: path.Join(gm.ContainerPatchesDir, fileName),
	}
}

func (gm *GitMapping) getPatchFileDescriptor(patchOpts git_repo.PatchOptions) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s.patch", objectToHashKey(patchOpts))

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gm.PatchesDir, fileName),
		ContainerFilePath: path.Join(gm.ContainerPatchesDir, fileName),
	}
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
