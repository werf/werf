package build

import (
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

	Name                 string
	As                   string
	Branch               string
	Tag                  string
	Commit               string
	To                   string
	RepoPath             string
	Cwd                  string
	Owner                string
	Group                string
	IncludePaths         []string
	ExcludePaths         []string
	StagesDependencies   map[string][]string
	Paramshash           string // TODO: method
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

func (ga *GitArtifact) applyPatchCommand(archiveType git_repo.ArchiveType, fromCommit, toCommit string) ([]string, error) {
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

	patchFile, err := ga.createPatchFile(fromCommit, toCommit)
	if err != nil {
		return nil, fmt.Errorf("cannot create patch between commits `%s` and `%s`: %s", fromCommit, toCommit, err)
	}

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

	patchOpts := git_repo.PatchOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	}

	anyChanges, err := ga.GitRepo().IsAnyChanges(patchOpts)
	if err != nil {
		return nil, err
	}

	if anyChanges {
		archiveType := git_repo.ArchiveType(
			stage.GetPrevStage().GetImage().
				GetLabels()[ga.getArchiveTypeLabelName()])

		// Verify archive-type not changed in to-commit repo state
		currentArchiveType, err := ga.GitRepo().ArchiveType(git_repo.ArchiveOptions{
			FilterOptions: ga.getRepoFilterOptions(),
			Commit:        toCommit,
		})
		if err != nil {
			return nil, err
		}
		if archiveType != currentArchiveType {
			return nil, fmt.Errorf("git repo `%s` archive type changed from `%s` to `%s`: reset cache manually and retry!", ga.GitRepo().String(), archiveType, currentArchiveType)
		}

		hasBinaryPatches, err := ga.GitRepo().HasBinaryPatches(patchOpts)
		if err != nil {
			return nil, err
		}

		if hasBinaryPatches {
			commands := make([]string, 0)

			commands = append(commands, fmt.Sprintf(
				"%s -rf %s",
				dappdeps.BaseBinPath("rm"),
				ga.To,
			))

			createArchiveCommands, err := ga.applyArchiveCommand(archiveType, toCommit)
			if err != nil {
				return nil, err
			}
			commands = append(commands, createArchiveCommands...)

			return commands, nil
		} else {
			return ga.applyPatchCommand(archiveType, fromCommit, toCommit)
		}
	}

	return nil, nil
}

func (ga *GitArtifact) applyArchiveCommand(archiveType git_repo.ArchiveType, commit string) ([]string, error) {
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

	// TODO: log create archive operation with time!
	archiveFile, err := ga.createArchiveFile(commit)
	if err != nil {
		return nil, fmt.Errorf("cannot create archive for commit `%s`: %s", commit, err)
	}

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

	anyEntries, err := ga.GitRepo().IsAnyEntries(git_repo.ArchiveOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		Commit:        commit,
	})

	if err != nil {
		return nil, err
	}

	if anyEntries {
		archiveType, err := ga.GitRepo().ArchiveType(git_repo.ArchiveOptions{
			FilterOptions: ga.getRepoFilterOptions(),
			Commit:        commit,
		})
		if err != nil {
			return nil, err
		}

		commands, err := ga.applyArchiveCommand(archiveType, commit)
		if err != nil {
			return nil, err
		}

		stage.GetImage().AddServiceChangeLabel(ga.getArchiveTypeLabelName(), string(archiveType))

		return commands, err
	}

	return nil, nil
}

func (ga *GitArtifact) getArchiveFileDescriptor(commit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s.tar", ga.Paramshash, commit)

	return &ContainerFileDescriptor{
		FilePath:          filepath.Join(ga.ArchivesDir, fileName),
		ContainerFilePath: filepath.Join(ga.ContainerArchivesDir, fileName),
	}
}

func (ga *GitArtifact) createArchiveFile(commit string) (*ContainerFileDescriptor, error) {
	fileDesc := ga.getArchiveFileDescriptor(commit)

	handler, err := fileDesc.Open(os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open archive file `%s`: %s", fileDesc.FilePath, err)
	}

	err = ga.GitRepo().CreateArchiveTar(handler, git_repo.ArchiveOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		Commit:        commit,
	})
	if err != nil {
		return nil, err
	}

	err = handler.Close()
	if err != nil {
		return nil, err
	}

	return fileDesc, nil
}

func (ga *GitArtifact) createPatchFile(fromCommit, toCommit string) (*ContainerFileDescriptor, error) {
	fileDesc := ga.getPatchFileDescriptor(fromCommit, toCommit)

	handler, err := fileDesc.Open(os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open archive file `%s`: %s", fileDesc.FilePath, err)
	}

	err = ga.GitRepo().CreatePatch(handler, git_repo.PatchOptions{
		FilterOptions: ga.getRepoFilterOptions(),
		FromCommit:    fromCommit,
		ToCommit:      toCommit,
	})
	if err != nil {
		return nil, err
	}

	err = handler.Close()
	if err != nil {
		return nil, err
	}

	return fileDesc, nil
}

func (ga *GitArtifact) getPatchFileDescriptor(fromCommit, toCommit string) *ContainerFileDescriptor {
	fileName := fmt.Sprintf("%s_%s_%s.patch", ga.Paramshash, fromCommit, toCommit)

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
	return fmt.Sprintf("dapp-git-%s-type", ga.Paramshash)
}
