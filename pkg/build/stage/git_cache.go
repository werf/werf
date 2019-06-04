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

func (s *GitCacheStage) GetDependencies(_ Conveyor, _, prevBuiltImage image.ImageInterface) (string, error) {
	var size int64
	for _, gitMapping := range s.gitMappings {
		commit := gitMapping.GetGitCommitFromImageLabels(prevBuiltImage)
		if commit == "" {
			return "", fmt.Errorf("invalid stage image: can not find git commit in stage image labels: delete stage image %s manually and retry the build", prevBuiltImage.Name())
		}

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

	return util.Sha256Hash(string(size / patchSizeStep)), nil
}
