package stage

import (
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

func (s *GitCacheStage) GetDependencies(_ Conveyor, prevImage image.ImageInterface) (string, error) {
	var size int64
	for _, gitMapping := range s.gitMappings {
		commit := gitMapping.GetGitCommitFromImageLabels(prevImage)
		if commit != "" {
			exist, err := gitMapping.GitRepo().IsCommitExists(commit)
			if err != nil {
				return "", err
			}

			if exist {
				patchSize, err := gitMapping.PatchSize(commit)
				if err != nil {
					return "", err
				}

				size += patchSize
			}
		}
	}

	return util.Sha256Hash(string(size / patchSizeStep)), nil
}
