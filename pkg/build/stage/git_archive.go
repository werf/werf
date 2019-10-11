package stage

import (
	"fmt"
	"sort"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
)

const GitArchiveResetCommitRegex = "(\\[werf reset\\])|(\\[reset werf\\])"

type NewGitArchiveStageOptions struct {
	ArchivesDir          string
	ScriptsDir           string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func NewGitArchiveStage(gitArchiveStageOptions *NewGitArchiveStageOptions, baseStageOptions *NewBaseStageOptions) *GitArchiveStage {
	s := &GitArchiveStage{
		ArchivesDir:          gitArchiveStageOptions.ArchivesDir,
		ScriptsDir:           gitArchiveStageOptions.ScriptsDir,
		ContainerArchivesDir: gitArchiveStageOptions.ContainerArchivesDir,
		ContainerScriptsDir:  gitArchiveStageOptions.ContainerScriptsDir,
	}
	s.GitStage = newGitStage(GitArchive, baseStageOptions)
	return s
}

type GitArchiveStage struct {
	*GitStage

	ArchivesDir          string
	ScriptsDir           string
	ContainerArchivesDir string
	ContainerScriptsDir  string
}

func (s *GitArchiveStage) GetDependencies(_ Conveyor, _, _ image.ImageInterface) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		args = append(args, gitMapping.GetParamshash())

		commit, err := gitMapping.GitRepo().FindCommitIdByMessage(GitArchiveResetCommitRegex)
		if err != nil {
			return "", err
		}

		args = append(args, commit)
	}

	sort.Strings(args)

	return util.Sha256Hash(args...), nil
}

func (s *GitArchiveStage) PrepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error {
	if err := s.GitStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.ApplyArchiveCommand(image); err != nil {
			return err
		}
	}

	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ArchivesDir, s.ContainerArchivesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

	return nil
}
