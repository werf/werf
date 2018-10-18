package build

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/git_repo"
)

type GitArtifact struct {
	LocalGitRepo  *git_repo.Local
	RemoteGitRepo *git_repo.Remote

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
	StagesDependencies map[string][]string

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

func (ga *GitArtifact) GitRepo() git_repo.GitRepo {
	if ga.LocalGitRepo != nil {
		return ga.LocalGitRepo
	} else if ga.RemoteGitRepo != nil {
		return ga.RemoteGitRepo
	}

	panic("GitRepo not initialized")
}

func (ga *GitArtifact) getRepoFilterOptions() git_repo.FilterOptions {
	return git_repo.FilterOptions{
		BasePath:     ga.RepoPath,
		IncludePaths: ga.IncludePaths,
		ExcludePaths: ga.ExcludePaths,
	}
}

func (ga *GitArtifact) IsLocal() bool {
	if ga.LocalGitRepo != nil {
		return true
	} else {
		return false
	}
}

func (ga *GitArtifact) LatestCommit() (string, error) {
	if ga.Commit != "" {
		fmt.Printf("Using specified commit `%s` of repository `%s`\n", ga.Commit, ga.GitRepo().String())
		return ga.Commit, nil
	}

	if ga.Tag != "" {
		return ga.GitRepo().LatestTagCommit(ga.Tag)
	}

	if ga.Branch != "" {
		return ga.GitRepo().LatestBranchCommit(ga.Branch)
	}

	return ga.GitRepo().HeadCommit()
}

func (ga *GitArtifact) applyPatchCommand(patchFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	commands := make([]string, 0)

	var applyPatchDirectory string

	switch archiveType {
	case git_repo.FileArchive:
		applyPatchDirectory = filepath.Dir(ga.To)
	case git_repo.DirectoryArchive:
		applyPatchDirectory = ga.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		dappdeps.BaseBinPath("install"),
		ga.makeCredentialsOpts(),
		applyPatchDirectory,
	))

	commands = append(commands, fmt.Sprintf(
		"%s %s apply --whitespace=nowarn --directory=\"%s\" --unsafe-paths %s",
		dappdeps.SudoCommand(ga.Owner, ga.Group),
		dappdeps.GitBin(),
		applyPatchDirectory,
		patchFile.ContainerFilePath,
	))

	return commands, nil
}

func (ga *GitArtifact) ApplyPatchCommand(stage Stage) ([]string, error) {
	fromCommit, err := stage.GetPrevStage().LayerCommit(ga)
	if err != nil {
		return nil, err
	}

	toCommit, err := stage.LayerCommit(ga)
	if err != nil {
		return nil, err
	}

	archiveType := git_repo.ArchiveType(
		stage.GetPrevStage().GetImage().
			GetLabels()[ga.getArchiveTypeLabelName()])

	patchOpts := git_repo.PatchOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}
	patch, err := ga.GitRepo().CreatePatch(patchOpts)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(patch.GetFilePath())

	if patch.IsEmpty() {
		return nil, nil
	}

	if patch.HasBinary() {
		patchPaths := patch.GetPaths()

		pathsListFile, err := ga.createPatchPathsListFile(patchPaths, fromCommit, toCommit)
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
			ga.To,
		))

		archiveOpts := git_repo.ArchiveOptions{
			FilterOptions: ga.getRepoFilterOptions(),
			Commit:        toCommit,
		}
		archive, err := ga.GitRepo().CreateArchive(archiveOpts)
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(archive.GetFilePath())

		if archive.IsEmpty() {
			return commands, nil
		}

		archiveFile, err := ga.createArchiveFile(archive, toCommit)
		if err != nil {
			return nil, fmt.Errorf("cannot create archive file: %s", err)
		}

		archiveType := archive.GetType()

		applyArchiveCommands, err := ga.applyArchiveCommand(archiveFile, archiveType)
		if err != nil {
			return nil, err
		}
		commands = append(commands, applyArchiveCommands...)

		return commands, nil
	}

	patchFile, err := ga.createPatchFile(patch, fromCommit, toCommit)
	if err != nil {
		return nil, fmt.Errorf("cannot create patch file: %s", err)
	}

	return ga.applyPatchCommand(patchFile, archiveType)
}

func (ga *GitArtifact) applyArchiveCommand(archiveFile *ContainerFileDescriptor, archiveType git_repo.ArchiveType) ([]string, error) {
	var unpackArchiveDirectory string
	commands := make([]string, 0)

	switch archiveType {
	case git_repo.FileArchive:
		unpackArchiveDirectory = filepath.Dir(ga.To)
	case git_repo.DirectoryArchive:
		unpackArchiveDirectory = ga.To
	default:
		return nil, fmt.Errorf("unknown archive type `%s`", archiveType)
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		dappdeps.BaseBinPath("install"),
		ga.makeCredentialsOpts(),
		unpackArchiveDirectory,
	))

	commands = append(commands, fmt.Sprintf(
		"%s %s -xf %s -C \"%s\"",
		dappdeps.SudoCommand(ga.Owner, ga.Group),
		dappdeps.BaseBinPath("tar"),
		archiveFile.ContainerFilePath,
		unpackArchiveDirectory,
	))

	return commands, nil
}

func (ga *GitArtifact) ApplyArchiveCommand(stage Stage) ([]string, error) {
	commit, err := stage.LayerCommit(ga)
	if err != nil {
		return nil, err
	}

	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		Commit:        commit,
	}
	archive, err := ga.GitRepo().CreateArchive(archiveOpts)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(archive.GetFilePath())

	if archive.IsEmpty() {
		return nil, nil
	}

	archiveFile, err := ga.createArchiveFile(archive, commit)
	if err != nil {
		return nil, fmt.Errorf("cannot create archive file: %s", err)
	}

	archiveType := archive.GetType()

	commands, err := ga.applyArchiveCommand(archiveFile, archiveType)
	if err != nil {
		return nil, err
	}

	stage.GetImage().AddServiceChangeLabel(ga.getArchiveTypeLabelName(), string(archiveType))

	return commands, err
}

func (ga *GitArtifact) StageDependenciesChecksum(stageName string) (string, error) {
	depsPaths := ga.StagesDependencies[stageName]
	if len(depsPaths) == 0 {
		return "", nil
	}

	commit, err := ga.LatestCommit()
	if err != nil {
		return "", fmt.Errorf("unable to get latest commit: %s", err)
	}

	opts := git_repo.ChecksumOptions{Paths: depsPaths, Commit: commit}

	checksum, err := ga.GitRepo().Checksum(opts)
	if err != nil {
		return "", err
	}

	for _, path := range checksum.GetNoMatchPaths() {
		fmt.Fprintf(os.Stderr, "WARN: stage `%s` dependency path `%s` have not been found in repo `%s`\n", stageName, path, ga.GitRepo().String())
	}

	return checksum.String(), nil
}

func (ga *GitArtifact) PatchSize(fromCommit string) (int64, error) {
	toCommit, err := ga.LatestCommit()
	if err != nil {
		return 0, fmt.Errorf("unable to get latest commit: %s", err)
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions:         ga.getRepoFilterOptions(),
		FromCommit:            fromCommit,
		ToCommit:              toCommit,
		WithEntireFileContext: true,
		WithBinary:            true,
	}
	patch, err := ga.GitRepo().CreatePatch(patchOpts)
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

func (ga *GitArtifact) GetFullName() string {
	if ga.Name != "" {
		return fmt.Sprintf("%s_%s", ga.GitRepo().GetName(), ga.Name)
	}
	return ga.GitRepo().GetName()
}

func (ga *GitArtifact) GetParamshash() string {
	var err error

	hash := sha256.New()

	parts := []string{ga.GetFullName(), ":::", ga.To, ":::", ga.Cwd}
	parts = append(parts, ":::")
	parts = append(parts, ga.IncludePaths...)
	parts = append(parts, ":::")
	parts = append(parts, ga.ExcludePaths...)
	parts = append(parts, ":::")
	parts = append(parts, ga.Owner)
	parts = append(parts, ":::")
	parts = append(parts, ga.Group)
	parts = append(parts, ":::")
	parts = append(parts, ga.Branch)
	parts = append(parts, ":::")
	parts = append(parts, ga.Tag)
	parts = append(parts, ":::")
	parts = append(parts, ga.Commit)

	for _, part := range parts {
		_, err = hash.Write([]byte(part))
		if err != nil {
			panic(fmt.Sprintf("error calculating sha256 of `%s`: %s", part, err))
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (ga *GitArtifact) IsPatchEmpty(stage Stage) (bool, error) {
	fromCommit, err := stage.GetPrevStage().LayerCommit(ga)
	if err != nil {
		return false, err
	}

	toCommit, err := stage.LayerCommit(ga)
	if err != nil {
		return false, err
	}

	patchOpts := git_repo.PatchOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}
	patch, err := ga.GitRepo().CreatePatch(patchOpts)
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(patch.GetFilePath())

	return patch.IsEmpty(), nil
}

func (ga *GitArtifact) IsEmpty() (bool, error) {
	commit, err := ga.LatestCommit()
	if err != nil {
		return false, fmt.Errorf("unable to get latest commit: %s", err)
	}

	archiveOpts := git_repo.ArchiveOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		Commit:        commit,
	}
	archive, err := ga.GitRepo().CreateArchive(archiveOpts)
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(archive.GetFilePath())

	return archive.IsEmpty(), nil
}

func (ga *GitArtifact) getArchiveFileDescriptor(commit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s.tar", ga.GetParamshash(), commit)

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(ga.ArchivesDir, fileName),
		ContainerFilePath: filepath.Join(ga.ContainerArchivesDir, fileName),
	}
}

func (ga *GitArtifact) createArchiveFile(archive git_repo.Archive, commit string) (*ContainerFileDescriptor, error) {
	fileDesc := ga.getArchiveFileDescriptor(commit)

	err := renameFile(archive.GetFilePath(), fileDesc.FilePath)
	if err != nil {
		return nil, err
	}

	return fileDesc, nil
}

func (ga *GitArtifact) createPatchPathsListFile(paths []string, fromCommit, toCommit string) (*ContainerFileDescriptor, error) {
	fileDesc := ga.getPatchPathsListFileDescriptor(fromCommit, toCommit)

	f, err := fileDesc.Open(os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("unable to open file `%s`: %s", fileDesc.FilePath, err)
	}

	fullPaths := make([]string, 0)
	for _, path := range paths {
		fullPaths = append(fullPaths, filepath.Join(ga.To, path))
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

func (ga *GitArtifact) createPatchFile(patch git_repo.Patch, fromCommit, toCommit string) (*ContainerFileDescriptor, error) {
	fileDesc := ga.getPatchFileDescriptor(fromCommit, toCommit)

	err := renameFile(patch.GetFilePath(), fileDesc.FilePath)
	if err != nil {
		return nil, err
	}

	return fileDesc, nil
}

func (ga *GitArtifact) getPatchPathsListFileDescriptor(fromCommit, toCommit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s_%s-paths-list", ga.GetParamshash(), fromCommit, toCommit)

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(ga.PatchesDir, fileName),
		ContainerFilePath: filepath.Join(ga.ContainerPatchesDir, fileName),
	}
}

func (ga *GitArtifact) getPatchFileDescriptor(fromCommit, toCommit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s_%s.patch", ga.GetParamshash(), fromCommit, toCommit)

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(ga.PatchesDir, fileName),
		ContainerFilePath: filepath.Join(ga.ContainerPatchesDir, fileName),
	}
}

func (ga *GitArtifact) makeCredentialsOpts() string {
	opts := make([]string, 0)

	if ga.Owner != "" {
		opts = append(opts, fmt.Sprintf("--owner=%s", ga.Owner))
	}
	if ga.Group != "" {
		opts = append(opts, fmt.Sprintf("--group=%s", ga.Group))
	}

	return strings.Join(opts, " ")
}

func (ga *GitArtifact) getArchiveTypeLabelName() string {
	return fmt.Sprintf("dapp-git-%s-type", ga.GetParamshash())
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
