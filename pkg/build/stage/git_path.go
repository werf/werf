package stage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/logger"
)

type GitPath struct {
	GitRepoInterface git_repo.GitRepo
	LocalGitRepo     *git_repo.Local
	RemoteGitRepo    *git_repo.Remote

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

func (gp *GitPath) GitRepo() git_repo.GitRepo {
	if gp.GitRepoInterface != nil {
		return gp.GitRepoInterface
	} else if gp.LocalGitRepo != nil {
		return gp.LocalGitRepo
	} else if gp.RemoteGitRepo != nil {
		return gp.RemoteGitRepo
	}

	panic("GitRepo not initialized")
}

func (gp *GitPath) getRepoFilterOptions() git_repo.FilterOptions {
	return git_repo.FilterOptions{
		BasePath:     gp.RepoPath,
		IncludePaths: gp.IncludePaths,
		ExcludePaths: gp.ExcludePaths,
	}
}

func (gp *GitPath) IsLocal() bool {
	if gp.LocalGitRepo != nil {
		return true
	} else {
		return false
	}
}

func (gp *GitPath) LatestCommit() (string, error) {
	if gp.Commit != "" {
		fmt.Printf("Using specified commit `%s` of repository `%s`\n", gp.Commit, gp.GitRepo().String())
		return gp.Commit, nil
	}

	if gp.Tag != "" {
		return gp.GitRepo().LatestTagCommit(gp.Tag)
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

func (gp *GitPath) applyPatchCommand(patchFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	commands := make([]string, 0)

	var applyPatchDirectory string

	switch archiveType {
	case git_repo.FileArchive:
		applyPatchDirectory = filepath.Dir(gp.To)
	case git_repo.DirectoryArchive:
		applyPatchDirectory = gp.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		dappdeps.BaseBinPath("install"),
		gp.makeCredentialsOpts(),
		applyPatchDirectory,
	))

	commands = append(commands, fmt.Sprintf(
		"%s %s apply --whitespace=nowarn --directory=\"%s\" --unsafe-paths %s",
		dappdeps.SudoCommand(gp.Owner, gp.Group),
		dappdeps.GitBin(),
		applyPatchDirectory,
		patchFile.ContainerFilePath,
	))

	return commands, nil
}

func (gp *GitPath) ApplyPatchCommand(prevBuiltImage, image image.Image) error {
	fromCommit, toCommit, err := gp.GetCommitsToPatch(prevBuiltImage)
	if err != nil {
		return err
	}

	commands, err := gp.baseApplyPatchCommand(fromCommit, toCommit, prevBuiltImage)
	if err != nil {
		return err
	}

	image.Container().AddRunCommands(commands...)

	gp.AddGitCommitToImageLabels(image, toCommit)

	return nil
}

func (gp *GitPath) GetCommitsToPatch(prevBuiltImage image.Image) (string, string, error) {
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

func (gp *GitPath) AddGitCommitToImageLabels(image image.Image, commit string) {
	image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
		gp.ImageGitCommitLabel(): commit,
	})
}

func (gp *GitPath) GetGitCommitFromImageLabels(builtImage image.Image) string {
	commit, ok := builtImage.Labels()[gp.ImageGitCommitLabel()]
	if !ok {
		return ""
	}

	return commit
}

func (gp *GitPath) ImageGitCommitLabel() string {
	return fmt.Sprintf("dapp-git-%s-commit", gp.GetParamshash())
}

func (gp *GitPath) baseApplyPatchCommand(fromCommit, toCommit string, prevBuiltImage image.Image) ([]string, error) {
	archiveType := git_repo.ArchiveType(prevBuiltImage.Labels()[gp.getArchiveTypeLabelName()])

	patchOpts := git_repo.PatchOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}
	patch, err := gp.GitRepo().CreatePatch(patchOpts)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(patch.GetFilePath())

	if patch.IsEmpty() {
		return nil, nil
	}

	if patch.HasBinary() {
		patchPaths := patch.GetPaths()

		pathsListFile, err := gp.createPatchPathsListFile(patchPaths, fromCommit, toCommit)
		if err != nil {
			return nil, fmt.Errorf("cannot create patch paths list file: %s", err)
		}

		commands := make([]string, 0)

		commands = append(commands, fmt.Sprintf(
			"%s --arg-file=%s --null %s --force",
			dappdeps.BaseBinPath("xargs"),
			pathsListFile.ContainerFilePath,
			dappdeps.BaseBinPath("rm"),
		))

		commands = append(commands, fmt.Sprintf(
			"%s %s -type d -empty -delete",
			dappdeps.BaseBinPath("find"),
			gp.To,
		))

		archiveOpts := git_repo.ArchiveOptions{
			FilterOptions: gp.getRepoFilterOptions(),
			Commit:        toCommit,
		}
		archive, err := gp.GitRepo().CreateArchive(archiveOpts)
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(archive.GetFilePath())

		if archive.IsEmpty() {
			return commands, nil
		}

		archiveFile, err := gp.createArchiveFile(archive, toCommit)
		if err != nil {
			return nil, fmt.Errorf("cannot create archive file: %s", err)
		}

		archiveType := archive.GetType()

		applyArchiveCommands, err := gp.applyArchiveCommand(archiveFile, archiveType)
		if err != nil {
			return nil, err
		}
		commands = append(commands, applyArchiveCommands...)

		return commands, nil
	}

	patchFile, err := gp.createPatchFile(patch, fromCommit, toCommit)
	if err != nil {
		return nil, fmt.Errorf("cannot create patch file: %s", err)
	}

	return gp.applyPatchCommand(patchFile, archiveType)
}

func (gp *GitPath) applyArchiveCommand(archiveFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	var unpackArchiveDirectory string
	commands := make([]string, 0)

	switch archiveType {
	case git_repo.FileArchive:
		unpackArchiveDirectory = filepath.Dir(gp.To)
	case git_repo.DirectoryArchive:
		unpackArchiveDirectory = gp.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		dappdeps.BaseBinPath("install"),
		gp.makeCredentialsOpts(),
		unpackArchiveDirectory,
	))

	commands = append(commands, fmt.Sprintf(
		"%s %s -xf %s -C \"%s\"",
		dappdeps.SudoCommand(gp.Owner, gp.Group),
		dappdeps.BaseBinPath("tar"),
		archiveFile.ContainerFilePath,
		unpackArchiveDirectory,
	))

	return commands, nil
}

func (gp *GitPath) ApplyArchiveCommand(image image.Image) error {
	commit, err := gp.LatestCommit()
	if err != nil {
		return err
	}

	commands, err := gp.baseApplyArchiveCommand(commit, image)
	if err != nil {
		return err
	}

	image.Container().AddRunCommands(commands...)

	gp.AddGitCommitToImageLabels(image, commit)

	return nil
}

func (gp *GitPath) baseApplyArchiveCommand(commit string, image image.Image) ([]string, error) {
	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		Commit:        commit,
	}
	archive, err := gp.GitRepo().CreateArchive(archiveOpts)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(archive.GetFilePath())

	if archive.IsEmpty() {
		return nil, nil
	}

	archiveFile, err := gp.createArchiveFile(archive, commit)
	if err != nil {
		return nil, fmt.Errorf("cannot create archive file: %s", err)
	}

	archiveType := archive.GetType()

	commands, err := gp.applyArchiveCommand(archiveFile, archiveType)
	if err != nil {
		return nil, err
	}

	image.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{gp.getArchiveTypeLabelName(): string(archiveType)})

	return commands, err
}

func (gp *GitPath) StageDependenciesChecksum(stageName StageName) (string, error) {
	depsPaths := gp.StagesDependencies[stageName]
	if len(depsPaths) == 0 {
		return "", nil
	}

	commit, err := gp.LatestCommit()
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit: %s", err)
	}

	opts := git_repo.ChecksumOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		Paths:         depsPaths,
		Commit:        commit,
	}

	checksum, err := gp.GitRepo().Checksum(opts)
	if err != nil {
		return "", err
	}

	for _, path := range checksum.GetNoMatchPaths() {
		logger.LogWarningF("WARNING: stage `%s` dependency path `%s` have not been found in repo `%s`\n", stageName, path, gp.GitRepo().String())
	}

	return checksum.String(), nil
}

func (gp *GitPath) PatchSize(fromCommit string) (int64, error) {
	toCommit, err := gp.LatestCommit()
	if err != nil {
		return 0, fmt.Errorf("unable to get latest commit: %s", err)
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions:         gp.getRepoFilterOptions(),
		FromCommit:            fromCommit,
		ToCommit:              toCommit,
		WithEntireFileContext: true,
		WithBinary:            true,
	}
	patch, err := gp.GitRepo().CreatePatch(patchOpts)
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(patch.GetFilePath())

	fileInfo, err := os.Stat(patch.GetFilePath())
	if err != nil {
		return 0, fmt.Errorf("unable to stat temporary patch file `%s`: %s", patch.GetFilePath(), err)
	}

	return fileInfo.Size(), nil
}

func (gp *GitPath) GetFullName() string {
	if gp.Name != "" {
		return fmt.Sprintf("%s_%s", gp.GitRepo().GetName(), gp.Name)
	}
	return gp.GitRepo().GetName()
}

func (gp *GitPath) GetParamshash() string {
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

func (gp *GitPath) IsPatchEmpty(prevBuiltImage image.Image) (bool, error) {
	fromCommit, toCommit, err := gp.GetCommitsToPatch(prevBuiltImage)
	if err != nil {
		return false, err
	}

	return gp.baseIsPatchEmpty(fromCommit, toCommit)
}

func (gp *GitPath) baseIsPatchEmpty(fromCommit, toCommit string) (bool, error) {
	patchOpts := git_repo.PatchOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}
	patch, err := gp.GitRepo().CreatePatch(patchOpts)
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(patch.GetFilePath())

	return patch.IsEmpty(), nil
}

func (gp *GitPath) IsEmpty() (bool, error) {
	commit, err := gp.LatestCommit()
	if err != nil {
		return false, fmt.Errorf("unable to get latest commit: %s", err)
	}

	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: gp.getRepoFilterOptions(),
		Commit:        commit,
	}
	archive, err := gp.GitRepo().CreateArchive(archiveOpts)
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(archive.GetFilePath())

	return archive.IsEmpty(), nil
}

func (gp *GitPath) getArchiveFileDescriptor(commit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s.tar", gp.GetParamshash(), commit)

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gp.ArchivesDir, fileName),
		ContainerFilePath: filepath.Join(gp.ContainerArchivesDir, fileName),
	}
}

func (gp *GitPath) createArchiveFile(archive git_repo.Archive, commit string) (*ContainerFileDescriptor, error) {
	fileDesc := gp.getArchiveFileDescriptor(commit)

	err := renameFile(archive.GetFilePath(), fileDesc.FilePath)
	if err != nil {
		return nil, err
	}

	return fileDesc, nil
}

func (gp *GitPath) createPatchPathsListFile(paths []string, fromCommit, toCommit string) (*ContainerFileDescriptor, error) {
	fileDesc := gp.getPatchPathsListFileDescriptor(fromCommit, toCommit)

	f, err := fileDesc.Open(os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("unable to open file `%s`: %s", fileDesc.FilePath, err)
	}

	fullPaths := make([]string, 0)
	for _, path := range paths {
		fullPaths = append(fullPaths, filepath.Join(gp.To, path))
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

func (gp *GitPath) createPatchFile(patch git_repo.Patch, fromCommit, toCommit string) (*ContainerFileDescriptor, error) {
	fileDesc := gp.getPatchFileDescriptor(fromCommit, toCommit)

	err := renameFile(patch.GetFilePath(), fileDesc.FilePath)
	if err != nil {
		return nil, err
	}

	return fileDesc, nil
}

func (gp *GitPath) getPatchPathsListFileDescriptor(fromCommit, toCommit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s_%s-paths-list", gp.GetParamshash(), fromCommit, toCommit)

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gp.PatchesDir, fileName),
		ContainerFilePath: filepath.Join(gp.ContainerPatchesDir, fileName),
	}
}

func (gp *GitPath) getPatchFileDescriptor(fromCommit, toCommit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s_%s.patch", gp.GetParamshash(), fromCommit, toCommit)

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(gp.PatchesDir, fileName),
		ContainerFilePath: filepath.Join(gp.ContainerPatchesDir, fileName),
	}
}

func (gp *GitPath) makeCredentialsOpts() string {
	opts := make([]string, 0)

	if gp.Owner != "" {
		opts = append(opts, fmt.Sprintf("--owner=%s", gp.Owner))
	}
	if gp.Group != "" {
		opts = append(opts, fmt.Sprintf("--group=%s", gp.Group))
	}

	return strings.Join(opts, " ")
}

func (gp *GitPath) getArchiveTypeLabelName() string {
	return fmt.Sprintf("dapp-git-%s-type", gp.GetParamshash())
}

func renameFile(fromPath, toPath string) error {
	var err error

	err = os.MkdirAll(filepath.Dir(toPath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.Rename(fromPath, toPath)
	if err != nil {
		return err
	}

	return nil
}
