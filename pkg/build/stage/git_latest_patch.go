package stage

import (
	"fmt"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

func NewGitLatestPatchStage(gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitLatestPatchStage {
	s := &GitLatestPatchStage{}
	s.GitPatchStage = newGitPatchStage(GitLatestPatch, gitPatchStageOptions, baseStageOptions)
	return s
}

type GitLatestPatchStage struct {
	*GitPatchStage
}

func (s *GitLatestPatchStage) IsEmpty(c Conveyor, prevBuiltImage container_runtime.ImageInterface) (bool, error) {
	if empty, err := s.GitPatchStage.IsEmpty(c, prevBuiltImage); err != nil {
		return false, err
	} else if empty {
		return true, nil
	}

	isEmpty := true
	for _, gitMapping := range s.gitMappings {
		commit, err := gitMapping.GetBaseCommitForPrevBuiltImage(c, prevBuiltImage)
		if err != nil {
			return false, fmt.Errorf("unable to get base commit for git mapping %s: %s", gitMapping.GitRepo().GetName(), err)
		}

		if exist, err := gitMapping.GitRepo().IsCommitExists(commit); err != nil {
			return false, fmt.Errorf("unable to check existance of commit %q in the repo %s: %s", commit, gitMapping.GitRepo().GetName(), err)
		} else if !exist {
			return false, fmt.Errorf("commit %q is not exist in the repo %s", commit, gitMapping.GitRepo().GetName())
		}

		if empty, err := gitMapping.IsPatchEmpty(c, prevBuiltImage); err != nil {
			return false, err
		} else if !empty {
			isEmpty = false
			break
		}
	}

	return isEmpty, nil
}

func (s *GitLatestPatchStage) GetDependencies(c Conveyor, _, prevBuiltImage container_runtime.ImageInterface) (string, error) {
	var args []string

	for _, gitMapping := range s.gitMappings {
		patchContent, err := gitMapping.GetPatchContent(c, prevBuiltImage)
		if err != nil {
			return "", fmt.Errorf("error getting patch between previous built image %s and current commit for git mapping %s: %s", prevBuiltImage.Name(), gitMapping.Name, err)
		}

		args = append(args, patchContent)
	}

	return util.Sha256Hash(args...), nil
}

func (s *GitLatestPatchStage) SelectSuitableStage(c Conveyor, stages []*image.StageDescription) (*image.StageDescription, error) {
	ancestorsImages, err := s.selectStagesAncestorsByGitMappings(c, stages)
	if err != nil {
		return nil, fmt.Errorf("unable to select cache images ancestors by git mappings: %s", err)
	}
	return s.selectStageByOldestCreationTimestamp(ancestorsImages)
}

func (s *GitLatestPatchStage) GetNextStageDependencies(c Conveyor) (string, error) {
	return s.BaseStage.getNextStageGitDependencies(c)
}
