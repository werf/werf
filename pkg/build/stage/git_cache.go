package stage

import (
	"fmt"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
)

const patchSizeStep = 1024 * 1024

func NewGitCacheStage(gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *GitCacheStage {
	s := &GitCacheStage{}
	s.GitPatchStage = newGitPatchStage(GitCache, gitPatchStageOptions, baseStageOptions)
	return s
}

type GitCacheStage struct {
	*GitPatchStage
}

func (s *GitCacheStage) IsEmpty(c Conveyor, prevBuiltImage image.ImageInterface) (bool, error) {
	if isEmptyBase, err := s.GitPatchStage.IsEmpty(c, prevBuiltImage); err != nil {
		return isEmptyBase, err
	} else if isEmptyBase {
		return true, err
	}

	patchSize, err := s.gitMappingsPatchSize(prevBuiltImage)
	if err != nil {
		return false, err
	}

	isEmpty := patchSize < patchSizeStep

	return isEmpty, nil
}

func (s *GitCacheStage) GetDependencies(_ Conveyor, _, prevBuiltImage image.ImageInterface) (string, error) {
	patchSize, err := s.gitMappingsPatchSize(prevBuiltImage)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(fmt.Sprintf("%d", patchSize/patchSizeStep)), nil
}

func (s *GitCacheStage) gitMappingsPatchSize(prevBuiltImage image.ImageInterface) (int64, error) {
	var size int64
	for _, gitMapping := range s.gitMappings {
		commit := gitMapping.GetGitCommitFromImageLabels(prevBuiltImage.Labels())
		if commit == "" {
			return 0, fmt.Errorf("invalid stage image: can not find git commit in stage image labels: delete stage image %s manually and retry the build", prevBuiltImage.Name())
		}

		exist, err := gitMapping.GitRepo().IsCommitExists(commit)
		if err != nil {
			return 0, err
		}

		if exist {
			patchSize, err := gitMapping.PatchSize(commit)
			if err != nil {
				return 0, err
			}

			size += patchSize
		} else {
			return 0, nil
		}
	}

	return size, nil
}
