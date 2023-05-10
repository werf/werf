package image

import (
	"sync"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/git_repo"
)

type Conveyor interface {
	stage.Conveyor

	GetImage(targetPlatform, name string) *Image
	GetOrCreateStageImage(name string, prevStageImage *stage.StageImage, stg stage.Interface, img *Image) *stage.StageImage

	GetForcedTargetPlatforms() []string
	GetTargetPlatforms() ([]string, error)
	GetImageTargetPlatforms(imageName string) ([]string, error)

	IsBaseImagesRepoIdsCacheExist(key string) bool
	GetBaseImagesRepoIdsCache(key string) string
	SetBaseImagesRepoIdsCache(key, value string)

	IsBaseImagesRepoErrCacheExist(key string) bool
	GetBaseImagesRepoErrCache(key string) error
	SetBaseImagesRepoErrCache(key string, err error)

	GetServiceRWMutex(service string) *sync.RWMutex

	SetRemoteGitRepo(key string, repo *git_repo.Remote)
	GetRemoteGitRepo(key string) *git_repo.Remote
}
