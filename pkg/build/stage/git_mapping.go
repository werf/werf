package stage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
	To                 string
	RepoPath           string
	Cwd                string
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

func (gp *GitMapping) GitRepo() git_repo.GitRepo {
	if gp.GitRepoInterface != nil {
		return gp.GitRepoInterface
	} else if gp.LocalGitRepo != nil {
		return gp.LocalGitRepo
	} else if gp.RemoteGitRepo != nil {
		return gp.RemoteGitRepo
	}

	panic("GitRepo not initialized")
}

func (gp *GitMapping) getOrCreateChecksum(opts git_repo.ChecksumOptions) (git_repo.Checksum, error) {
	if _, hasKey := gp.GitRepoCache.Checksums[objectToHashKey(opts)]; !hasKey {
		checksum, err := gp.GitRepo().Checksum(opts)
		if err != nil {
			return nil, err
		}
		gp.GitRepoCache.Checksums[objectToHashKey(opts)] = checksum
	}
	return gp.GitRepoCache.Checksums[objectToHashKey(opts)], nil
}

func (gp *GitMapping) getOrCreateArchive(opts git_repo.ArchiveOptions) (git_repo.Archive, error) {
	if _, hasKey := gp.GitRepoCache.Archives[objectToHashKey(opts)]; !hasKey {
		archive, err := gp.createArchive(opts)
		if err != nil {
			return nil, err
		}
		gp.GitRepoCache.Archives[objectToHashKey(opts)] = archive
	}
	return gp.GitRepoCache.Archives[objectToHashKey(opts)], nil
}

func (gp *GitMapping) createArchive(opts git_repo.ArchiveOptions) (git_repo.Archive, error) {
	var res git_repo.Archive

	err := logboek.LogProcess(fmt.Sprintf("Creating archive for commit %s of %s git mapping %s", opts.Commit, gp.GitRepo().GetName(), gp.Cwd), logboek.LogProcessOptions{}, func() error {
		archive, err := gp.GitRepo().CreateArchive(opts)
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

func (gp *GitMapping) getOrCreatePatch(opts git_repo.PatchOptions) (git_repo.Patch, error) {
	if _, hasKey := gp.GitRepoCache.Patches[objectToHashKey(opts)]; !hasKey {
		patch, err := gp.createPatch(opts)
		if err != nil {
			return nil, err
		}
		gp.GitRepoCache.Patches[objectToHashKey(opts)] = patch
	}
	return gp.GitRepoCache.Patches[objectToHashKey(opts)], nil
}

func (gp *GitMapping) createPatch(opts git_repo.PatchOptions) (git_repo.Patch, error) {
	var res git_repo.Patch

	logProcessMsg := fmt.Sprintf("Creating patch %s..%s for %s git mapping %s", opts.FromCommit, opts.ToCommit, gp.GitRepo().GetName(), gp.Cwd)
	err := logboek.LogProcess(logProcessMsg, logboek.LogProcessOptions{}, func() error {
		patch, err := gp.GitRepo().CreatePatch(opts)
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

func (gp *GitMapping) getRepoFilterOptions() git_repo.FilterOptions {
	return git_repo.FilterOptions{
		BasePath:     gp.RepoPath,
		IncludePaths: gp.IncludePaths,
		ExcludePaths: gp.ExcludePaths,
	}
}

func (gp *GitMapping) IsLocal() bool {
	if gp.LocalGitRepo != nil {
		return true
	} else {
		return false
	}
}

func (gp *GitMapping) LatestCommit() (string, error) {
	if gp.Commit != "" {
		return gp.Commit, nil
	}

	if gp.Tag != "" {
		return gp.GitRepo().TagCommit(gp.Tag)
	}

	if gp.Branch != "" {
		return gp.GitRepo().LatestBranchCommit(gp.Branch)
	}

	commit, err := gp.GitRepo().HeadCommit()
	if err != nil {
		return "", err
	}

	return commit, nil
}

func (gp *GitMapping) applyPatchCommand(patchFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	commands := make([]string, 0)

	var applyPatchDirectory string

	switch archiveType {
	case git_repo.FileArchive:
		applyPatchDirectory = path.Dir(gp.To)
	case git_repo.DirectoryArchive:
		applyPatchDirectory = gp.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		stapel.InstallBinPath(),
		gp.makeCredentialsOpts(),
		applyPatchDirectory,
	))

	gitCommand := fmt.Sprintf(
		"%s %s apply --whitespace=nowarn --directory=\"%s\" --unsafe-paths %s",
		stapel.OptionalSudoCommand(gp.Owner, gp.Group),
		stapel.GitBinPath(),
		applyPatchDirectory,
		patchFile.ContainerFilePath,
	)

	commands = append(commands, strings.TrimLeft(gitCommand, " "))

	return commands, nil
}

func (gp *GitMapping) ApplyPatchCommand(prevBuiltImage, image image.ImageInterface) error {
	fromCommit, toCommit, err := gp.GetCommitsToPatch(prevBuiltImage)
	if err != nil {
		return err
	}

	commands, err := gp.baseApplyPatchCommand(fromCommit, toCommit, prevBuiltImage)
	if err != nil {
		return err
	}

	if err := gp.applyScript(image, commands); err != nil {
		return err
	}

	gp.AddGitCommitToImageLabels(image, toCommit)

	return nil
}

func (gp *GitMapping) GetCommitsToPatch(prevBuiltImage image.ImageInterface) (string, string, error) {
	fromCommit := gp.GetGitCommitFromImageLabels(prevBuiltImage)
	if fromCommit == "" {
		panic("Commit should be in prev built image labels!")
	}

	toCommit, err := gp.LatestCommit()
	if err != nil {
		return "", "", err
	}

	return fromCommit, toCommit, nil
}

func (gp *GitMapping) AddGitCommitToImageLabels(image image.ImageInterface, commit string) {
	image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
		gp.ImageGitCommitLabel(): commit,
	})
}

func (gp *GitMapping) GetGitCommitFromImageLabels(builtImage image.ImageInterface) string {
	commit, ok := builtImage.Labels()[gp.ImageGitCommitLabel()]
	if !ok {
		return ""
	}

	return commit
}

func (gp *GitMapping) ImageGitCommitLabel() string {
	return fmt.Sprintf("werf-git-%s-commit", gp.GetParamshash())
}

func (gp *GitMapping) baseApplyPatchCommand(fromCommit, toCommit string, prevBuiltImage image.ImageInterface) ([]string, error) {
	archiveType := git_repo.ArchiveType(prevBuiltImage.Labels()[gp.getArchiveTypeLabelName()])

	patchOpts := git_repo.PatchOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}

	patch, err := gp.getOrCreatePatch(patchOpts)
	if err != nil {
		return nil, err
	}

	if patch.IsEmpty() {
		return nil, nil
	}

	if patch.HasBinary() {
		pathsListFile, err := gp.preparePatchPathsListFile(patchOpts, patch)
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
					quoteShellArg(filepath.Join(gp.To, targetRelDir)),
				))
			}
		}

		archiveOpts := git_repo.ArchiveOptions{
			FilterOptions: gp.getRepoFilterOptions(),
			Commit:        toCommit,
		}

		archive, err := gp.getOrCreateArchive(archiveOpts)
		if err != nil {
			return nil, err
		}

		if archive.IsEmpty() {
			commands = append(commands, rmEmptyChangedDirsCommands...)
			return commands, nil
		}

		archiveFile, err := gp.prepareArchiveFile(archiveOpts, archive)
		if err != nil {
			return nil, fmt.Errorf("cannot prepare archive file: %s", err)
		}

		archiveType := archive.GetType()

		applyArchiveCommands, err := gp.applyArchiveCommand(archiveFile, archiveType)
		if err != nil {
			return nil, err
		}
		commands = append(commands, applyArchiveCommands...)

		commands = append(commands, rmEmptyChangedDirsCommands...)

		return commands, nil
	}

	patchFile, err := gp.preparePatchFile(patchOpts, patch)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare patch file: %s", err)
	}

	return gp.applyPatchCommand(patchFile, archiveType)
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

func (gp *GitMapping) applyArchiveCommand(archiveFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	var unpackArchiveDirectory string
	commands := make([]string, 0)

	switch archiveType {
	case git_repo.FileArchive:
		unpackArchiveDirectory = path.Dir(gp.To)
	case git_repo.DirectoryArchive:
		unpackArchiveDirectory = gp.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		stapel.InstallBinPath(),
		gp.makeCredentialsOpts(),
		unpackArchiveDirectory,
	))

	tarCommand := fmt.Sprintf(
		"%s %s -xf %s -C \"%s\"",
		stapel.OptionalSudoCommand(gp.Owner, gp.Group),
		stapel.TarBinPath(),
		archiveFile.ContainerFilePath,
		unpackArchiveDirectory,
	)

	commands = append(commands, strings.TrimLeft(tarCommand, " "))

	return commands, nil
}

func (gp *GitMapping) ApplyArchiveCommand(image image.ImageInterface) error {
	commit, err := gp.LatestCommit()
	if err != nil {
		return err
	}

	commands, err := gp.baseApplyArchiveCommand(commit, image)
	if err != nil {
		return err
	}

	if err := gp.applyScript(image, commands); err != nil {
		return err
	}

	gp.AddGitCommitToImageLabels(image, commit)

	return nil
}

func (gp *GitMapping) applyScript(image image.ImageInterface, commands []string) error {
	stageHostTmpScriptFilePath := filepath.Join(gp.ScriptsDir, gp.GetParamshash())
	containerTmpScriptFilePath := path.Join(gp.ContainerScriptsDir, gp.GetParamshash())

	if err := stapel.CreateScript(stageHostTmpScriptFilePath, commands); err != nil {
		return err
	}

	image.Container().AddServiceRunCommands(containerTmpScriptFilePath)

	return nil
}

func (gp *GitMapping) baseApplyArchiveCommand(commit string, image image.ImageInterface) ([]string, error) {
	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		Commit:        commit,
	}

	archive, err := gp.getOrCreateArchive(archiveOpts)
	if err != nil {
		return nil, err
	}

	if archive.IsEmpty() {
		return nil, nil
	}

	archiveFile, err := gp.prepareArchiveFile(archiveOpts, archive)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare archive file: %s", err)
	}

	archiveType := archive.GetType()

	commands, err := gp.applyArchiveCommand(archiveFile, archiveType)
	if err != nil {
		return nil, err
	}

	image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{gp.getArchiveTypeLabelName(): string(archiveType)})

	return commands, err
}

func (gp *GitMapping) StageDependenciesChecksum(stageName StageName) (string, error) {
	depsPaths := gp.StagesDependencies[stageName]
	if len(depsPaths) == 0 {
		return "", nil
	}

	commit, err := gp.LatestCommit()
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit: %s", err)
	}

	checksumOpts := git_repo.ChecksumOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		Paths:         depsPaths,
		Commit:        commit,
	}

	checksum, err := gp.getOrCreateChecksum(checksumOpts)
	if err != nil {
		return "", err
	}

	for _, p := range checksum.GetNoMatchPaths() {
		logboek.LogErrorF("WARNING: stage %s dependency path %s have not been found in %s git\n", stageName, p,
			gp.GitRepo().GetName())
	}

	return checksum.String(), nil
}

func (gp *GitMapping) PatchSize(fromCommit string) (int64, error) {
	toCommit, err := gp.LatestCommit()
	if err != nil {
		return 0, fmt.Errorf("unable to get latest commit: %s", err)
	}

	if fromCommit == toCommit {
		return 0, nil
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions:         gp.getRepoFilterOptions(),
		FromCommit:            fromCommit,
		ToCommit:              toCommit,
		WithEntireFileContext: true,
		WithBinary:            true,
	}

	patch, err := gp.getOrCreatePatch(patchOpts)
	if err != nil {
		return 0, err
	}

	fileInfo, err := os.Stat(patch.GetFilePath())
	if err != nil {
		return 0, fmt.Errorf("unable to stat temporary patch file `%s`: %s", patch.GetFilePath(), err)
	}

	return fileInfo.Size(), nil
}

func (gp *GitMapping) GetFullName() string {
	if gp.Name != "" {
		return fmt.Sprintf("%s_%s", gp.GitRepo().GetName(), gp.Name)
	}
	return gp.GitRepo().GetName()
}

func (gp *GitMapping) GetParamshash() string {
	var err error

	hash := sha256.New()

	parts := []string{gp.GetFullName(), ":::", gp.To, ":::", gp.Cwd}
	parts = append(parts, ":::")
	parts = append(parts, gp.IncludePaths...)
	parts = append(parts, ":::")
	parts = append(parts, gp.ExcludePaths...)
	parts = append(parts, ":::")
	parts = append(parts, gp.Owner)
	parts = append(parts, ":::")
	parts = append(parts, gp.Group)
	parts = append(parts, ":::")
	parts = append(parts, gp.Branch)
	parts = append(parts, ":::")
	parts = append(parts, gp.Tag)
	parts = append(parts, ":::")
	parts = append(parts, gp.Commit)

	for _, part := range parts {
		_, err = hash.Write([]byte(part))
		if err != nil {
			panic(fmt.Sprintf("error calculating sha256 of `%s`: %s", part, err))
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (gp *GitMapping) IsPatchEmpty(prevBuiltImage image.ImageInterface) (bool, error) {
	fromCommit, toCommit, err := gp.GetCommitsToPatch(prevBuiltImage)
	if err != nil {
		return false, err
	}

	return gp.baseIsPatchEmpty(fromCommit, toCommit)
}

func (gp *GitMapping) baseIsPatchEmpty(fromCommit, toCommit string) (bool, error) {
	if fromCommit == toCommit {
		return true, nil
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}

	patch, err := gp.getOrCreatePatch(patchOpts)
	if err != nil {
		return false, err
	}

	return patch.IsEmpty(), nil
}

func (gp *GitMapping) IsEmpty() (bool, error) {
	commit, err := gp.LatestCommit()
	if err != nil {
		return false, fmt.Errorf("unable to get latest commit: %s", err)
	}

	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		Commit:        commit,
	}

	archive, err := gp.getOrCreateArchive(archiveOpts)
	if err != nil {
		return false, err
	}

	return archive.IsEmpty(), nil
}

func (gp *GitMapping) getArchiveFileDescriptor(archiveOpts git_repo.ArchiveOptions) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s.tar", objectToHashKey(archiveOpts))

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gp.ArchivesDir, fileName),
		ContainerFilePath: path.Join(gp.ContainerArchivesDir, fileName),
	}
}

func (gp *GitMapping) prepareArchiveFile(archiveOpts git_repo.ArchiveOptions, archive git_repo.Archive) (*ContainerFileDescriptor, error) {
	fileDesc := gp.getArchiveFileDescriptor(archiveOpts)

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

func (gp *GitMapping) preparePatchPathsListFile(patchOpts git_repo.PatchOptions, patch git_repo.Patch) (*ContainerFileDescriptor, error) {
	fileDesc := gp.getPatchPathsListFileDescriptor(patchOpts)

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
		fullPaths = append(fullPaths, path.Join(gp.To, p))
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

func (gp *GitMapping) preparePatchFile(patchOpts git_repo.PatchOptions, patch git_repo.Patch) (*ContainerFileDescriptor, error) {
	fileDesc := gp.getPatchFileDescriptor(patchOpts)

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

func (gp *GitMapping) getPatchPathsListFileDescriptor(patchOpts git_repo.PatchOptions) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s.paths_list", objectToHashKey(patchOpts))

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gp.PatchesDir, fileName),
		ContainerFilePath: path.Join(gp.ContainerPatchesDir, fileName),
	}
}

func (gp *GitMapping) getPatchFileDescriptor(patchOpts git_repo.PatchOptions) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s.patch", objectToHashKey(patchOpts))

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gp.PatchesDir, fileName),
		ContainerFilePath: path.Join(gp.ContainerPatchesDir, fileName),
	}
}

func (gp *GitMapping) makeCredentialsOpts() string {
	opts := make([]string, 0)

	if gp.Owner != "" {
		opts = append(opts, fmt.Sprintf("--owner=%s", gp.Owner))
	}
	if gp.Group != "" {
		opts = append(opts, fmt.Sprintf("--group=%s", gp.Group))
	}

	return strings.Join(opts, " ")
}

func (gp *GitMapping) getArchiveTypeLabelName() string {
	return fmt.Sprintf("werf-git-%s-type", gp.GetParamshash())
}
