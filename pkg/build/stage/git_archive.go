package stage

import (
	"context"
	"fmt"
	"sort"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/image"
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

func (s *GitArchiveStage) SelectSuitableStage(ctx context.Context, c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error) {
	ancestorsStages, err := s.selectStagesAncestorsByGitMappings(ctx, c, stages)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsStages)
}

// TODO: 1.3 add git mapping type (dir, file, ...) to gitArchive stage digest
func (s *GitArchiveStage) GetDependencies(ctx context.Context, c Conveyor, _, _ container_runtime.ImageInterface) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		if gitMapping.LocalGitRepo != nil {
			if err := c.GiterminismManager().Inspector().InspectBuildContextFiles(ctx, gitMapping.getPathMatcher()); err != nil {
				return "", err
			}
		}

		args = append(args, gitMapping.GetParamshash())
	}

	sort.Strings(args)

	return util.Sha256Hash(args...), nil
}

func (s *GitArchiveStage) GetNextStageDependencies(ctx context.Context, c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(ctx, c)
}

func (s *GitArchiveStage) PrepareImage(ctx context.Context, c Conveyor, prevBuiltImage, image container_runtime.ImageInterface) error {
	if err := s.GitStage.PrepareImage(ctx, c, prevBuiltImage, image); err != nil {
		return err
	}

	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.ApplyArchiveCommand(ctx, c, image); err != nil {
			return err
		}
	}

	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", git_repo.CommonGitDataManager.GetArchivesCacheDir(), s.ContainerArchivesDir))
	image.Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))

	return nil
}
