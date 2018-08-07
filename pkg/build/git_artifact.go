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

	Name                string
	As                  string
	Branch              string
	Tag                 string
	Commit              string
	To                  string
	RepoPath            string
	Cwd                 string
	Owner               string
	Group               string
	IncludePaths        []string
	ExcludePaths        []string
	StagesDependencies  map[string][]string
	Paramshash          string // TODO: method
	PatchesDir          string
	ContainerPatchesDir string
}

func (ga *GitArtifact) GitRepo() git_repo.GitRepo {
	if ga.LocalGitRepo != nil {
		return ga.LocalGitRepo
	} else if ga.RemoteGitRepo != nil {
		return ga.RemoteGitRepo
	}

	panic("GitRepo not initialized")
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

func (ga *GitArtifact) patchCommands(applyPatchDirectory string, fromCommit, toCommit string) ([]string, error) {
	commands := make([]string, 0)

	commands = append(commands, fmt.Sprintf(
		"%s %s -d \"%s\"",
		dappdeps.BaseBinPath("install"),
		ga.makeCredentialsOpts(),
		applyPatchDirectory,
	))

	patch, err := ga.GitRepo().Diff(ga.RepoPath, fromCommit, toCommit, ga.IncludePaths, ga.ExcludePaths)
	if err != nil {
		return nil, err
	}

	containerPatchFilePath, err := ga.makePatchFile(patch, fromCommit, toCommit)
	if err != nil {
		return nil, err
	}

	commands = append(commands, fmt.Sprintf(
		"%s %s apply --whitespace=nowarn --directory=\"%s\" --unsafe-paths %s",
		dappdeps.SudoCommand(ga.Owner, ga.Group),
		dappdeps.GitBin(),
		applyPatchDirectory,
		containerPatchFilePath,
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

	anyChanges, err := ga.GitRepo().IsAnyChanges(ga.RepoPath, fromCommit, toCommit, ga.IncludePaths, ga.ExcludePaths)
	if err != nil {
		return nil, err
	}

	commands := make([]string, 0)

	if anyChanges {
		var applyPatchDirectory string

		switch archiveType := ga.ArchiveType(stage); archiveType {
		case "file":
			applyPatchDirectory = filepath.Dir(ga.To)
		case "directory":
			applyPatchDirectory = ga.To
		default:
			return []string{}, fmt.Errorf("unknown archive type `%s`", archiveType)
		}

		patchCommands, err := ga.patchCommands(applyPatchDirectory, fromCommit, toCommit)
		if err != nil {
			return nil, err
		}
		commands = append(commands, patchCommands...)
	}

	return commands, nil
}

func (ga *GitArtifact) makePatchFile(patch string, fromCommit, toCommit string) (string, error) {
	fileName := fmt.Sprintf("%s_%s_%s.patch", ga.Paramshash, fromCommit, toCommit)

	filePath := filepath.Join(ga.PatchesDir, fileName)
	containerFilePath := filepath.Join(ga.ContainerPatchesDir, fileName)

	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("cannot create patch file directory: %s", err)
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return "", fmt.Errorf("cannot create patch file: %s", err)
	}

	_, err = f.Write([]byte(patch))
	if err != nil {
		return "", fmt.Errorf("cannot write patch file data: %s", err)
	}

	return containerFilePath, nil
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

func (ga *GitArtifact) ArchiveType(stage Stage) string {
	return stage.GetPrevStage().GetImage().GetLabels()[fmt.Sprintf("dapp-git-%s-type", ga.Paramshash)]
}

func (ga *GitArtifact) IsAnyChanges(fromCommit, toCommit string) (bool, error) {
	return true, nil
}

func (ga *GitArtifact) Diff(fromCommit, toCommit string, paths []string) error {
	return nil
}
