package stage

import (
	"fmt"
	"sort"

	"github.com/werf/werf/pkg/image"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/util"
)

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

func (s *GitArchiveStage) SelectSuitableStage(c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error) {
	ancestorsStages, err := s.selectStagesAncestorsByGitMappings(c, stages)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsStages)
}

func (s *GitArchiveStage) GetDependencies(_ Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		args = append(args, gitMapping.GetParamshash())
	}

	sort.Strings(args)

	return util.Sha256Hash(args...), nil
}

func (s *GitArchiveStage) GetNextStageDependencies(c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(c)
}

func (s *GitArchiveStage) PrepareImage(c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.GitStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return err
	}

	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.ApplyArchiveCommand(c, image); err != nil {
			return err
		}
	}

	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ArchivesDir, s.ContainerArchivesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

	return nil
}
