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

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/util"
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

	err := logboek.LogProcess(fmt.Sprintf("Creating archive for commit %s of %s git mapping %s", opts.Commit, gm.GitRepo().GetName(), gm.Add), logboek.LogProcessOptions{}, func() error {
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
	err := logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
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

func (gm *GitMapping) LatestCommit() (string, error) {
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

func (gm *GitMapping) ApplyPatchCommand(prevBuiltImage, image image.ImageInterface) error {
	fromCommit, toCommit, err := gm.GetCommitsToPatch(prevBuiltImage)
	if err != nil {
		return err
	}

	commands, err := gm.baseApplyPatchCommand(fromCommit, toCommit, prevBuiltImage)
	if err != nil {
		return err
	}

	if err := gm.applyScript(image, commands); err != nil {
		return err
	}

	gm.AddGitCommitToImageLabels(image, toCommit)

	return nil
}

func (gm *GitMapping) GetCommitsToPatch(prevBuiltImage image.ImageInterface) (string, string, error) {
	fromCommit := gm.GetGitCommitFromImageLabels(prevBuiltImage.Labels())
	if fromCommit == "" {
		panic("Commit should be in prev built image labels!")
	}

	toCommit, err := gm.LatestCommit()
	if err != nil {
		return "", "", err
	}

	return fromCommit, toCommit, nil
}

func (gm *GitMapping) AddGitCommitToImageLabels(image image.ImageInterface, commit string) {
	image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
		gm.ImageGitCommitLabel(): commit,
	})
}

func (gm *GitMapping) GetGitCommitFromImageLabels(labels map[string]string) string {
	commit, ok := labels[gm.ImageGitCommitLabel()]
	if !ok {
		return ""
	}

	return commit
}

func (gm *GitMapping) ImageGitCommitLabel() string {
	return fmt.Sprintf("werf-git-%s-commit", gm.GetParamshash())
}

func (gm *GitMapping) baseApplyPatchCommand(fromCommit, toCommit string, prevBuiltImage image.ImageInterface) ([]string, error) {
	archiveType := git_repo.ArchiveType(prevBuiltImage.Labels()[gm.getArchiveTypeLabelName()])

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

func (gm *GitMapping) ApplyArchiveCommand(image image.ImageInterface) error {
	commit, err := gm.LatestCommit()
	if err != nil {
		return err
	}

	commands, err := gm.baseApplyArchiveCommand(commit, image)
	if err != nil {
		return err
	}

	if err := gm.applyScript(image, commands); err != nil {
		return err
	}

	gm.AddGitCommitToImageLabels(image, commit)

	return nil
}

func (gm *GitMapping) applyScript(image image.ImageInterface, commands []string) error {
	stageHostTmpScriptFilePath := filepath.Join(gm.ScriptsDir, gm.GetParamshash())
	containerTmpScriptFilePath := path.Join(gm.ContainerScriptsDir, gm.GetParamshash())

	if err := stapel.CreateScript(stageHostTmpScriptFilePath, commands); err != nil {
		return err
	}

	image.Container().AddServiceRunCommands(containerTmpScriptFilePath)

	return nil
}

func (gm *GitMapping) baseApplyArchiveCommand(commit string, image image.ImageInterface) ([]string, error) {
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

func (gm *GitMapping) StageDependenciesChecksum(stageName StageName) (string, error) {
	depsPaths := gm.StagesDependencies[stageName]
	if len(depsPaths) == 0 {
		return "", nil
	}

	commit, err := gm.LatestCommit()
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit: %s", err)
	}

	checksumOpts := git_repo.ChecksumOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		Paths:         depsPaths,
		Commit:        commit,
	}

	checksum, err := gm.getOrCreateChecksum(checksumOpts)
	if err != nil {
		return "", err
	}

	for _, p := range checksum.GetNoMatchPaths() {
		logboek.LogWarnF(
			"WARNING: stage %s dependency path %s have not been found in %s git\n",
			stageName, p, gm.GitRepo().GetName(),
		)
	}

	return checksum.String(), nil
}

func (gm *GitMapping) PatchSize(fromCommit string) (int64, error) {
	toCommit, err := gm.LatestCommit()
	if err != nil {
		return 0, fmt.Errorf("unable to get latest commit: %s", err)
	}

	if fromCommit == toCommit {
		return 0, nil
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions:         gm.getRepoFilterOptions(),
		FromCommit:            fromCommit,
		ToCommit:              toCommit,
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

	parts := []string{gm.GetFullName(), ":::", gm.To, ":::", gm.Add}
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

func (gm *GitMapping) GetPatchContent(prevBuiltImage image.ImageInterface) (string, error) {
	fromCommit, toCommit, err := gm.GetCommitsToPatch(prevBuiltImage)
	if err != nil {
		return "", err
	}
	if fromCommit == toCommit {
		return "", nil
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
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

func (gm *GitMapping) IsPatchEmpty(prevBuiltImage image.ImageInterface) (bool, error) {
	fromCommit, toCommit, err := gm.GetCommitsToPatch(prevBuiltImage)
	if err != nil {
		return false, err
	}

	return gm.baseIsPatchEmpty(fromCommit, toCommit)
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

func (gm *GitMapping) IsEmpty() (bool, error) {
	commit, err := gm.LatestCommit()
	if err != nil {
		return false, fmt.Errorf("unable to get latest commit: %s", err)
	}

	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: gm.getRepoFilterOptions(),
		Commit:        commit,
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
