package stage

import (
	"context"
	"fmt"
	"sort"

	"github.com/werf/werf/pkg/container_backend"
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

func NewGitArchiveStage(gitArchiveStageOptions *NewGitArchiveStageOptions, baseStageOptions *BaseStageOptions) *GitArchiveStage {
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
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %w", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsStages)
}

// TODO: 1.3 add git mapping type (dir, file, ...) to gitArchive stage digest
func (s *GitArchiveStage) GetDependencies(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevImage, prevBuiltImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) (string, error) {
	var args []string
	for _, gitMapping := range s.gitMappings {
		if gitMapping.IsLocal() {
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

func (s *GitArchiveStage) PrepareImage(ctx context.Context, c Conveyor, cb container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := s.GitStage.PrepareImage(ctx, c, cb, prevBuiltImage, stageImage, nil); err != nil {
		return err
	}

	for _, gitMapping := range s.gitMappings {
		if err := gitMapping.PrepareArchiveForImage(ctx, c, cb, stageImage); err != nil {
			return fmt.Errorf("unable to prepare git mapping %s for image stage: %w", gitMapping.Name, err)
		}
	}

	if c.UseLegacyStapelBuilder(cb) {
		stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", git_repo.CommonGitDataManager.GetArchivesCacheDir(), s.ContainerArchivesDir))
		stageImage.Builder.LegacyStapelStageBuilder().Container().RunOptions().AddVolume(fmt.Sprintf("%s:%s:ro", s.ScriptsDir, s.ContainerScriptsDir))
	}

	return nil
}

func (s *GitArchiveStage) IsEmpty(ctx context.Context, c Conveyor, stageImage *StageImage) (bool, error) {
	for _, gitMapping := range s.gitMappings {
		isGitMappingEmpty, err := gitMapping.isEmpty(ctx, c)
		if err != nil {
			return false, fmt.Errorf("error checking git mapping emptiness: %w", err)
		}
		if isGitMappingEmpty {
			return false, fmt.Errorf(`"git.add: /%s" in werf.yaml matches no files. git.add requires at least one file matched by it. Fix and retry`, gitMapping.Add)
		}
	}

	return s.GitStage.IsEmpty(ctx, c, stageImage)
}
