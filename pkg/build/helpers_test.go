package build

import (
	"context"

	"github.com/onsi/gomega"

	buildImage "github.com/werf/werf/v3/pkg/build/image"
	"github.com/werf/werf/v3/pkg/build/stage"
	"github.com/werf/werf/v3/pkg/container_backend"
	imagePkg "github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
)

func newImage(name string, baseImageType buildImage.BaseImageType, opts buildImage.ImageOptions) *buildImage.Image {
	img, err := buildImage.NewImage(context.Background(), "linux/amd64", name, baseImageType, opts)
	gomega.Expect(err).To(gomega.Succeed())
	return img
}

type contentDependenciesStub struct {
	*stage.BaseStage
	deps string
	err  error
}

var _ stage.Interface = (*contentDependenciesStub)(nil)

func newContentDependenciesStub(name stage.StageName, deps string) *contentDependenciesStub {
	return &contentDependenciesStub{BaseStage: stage.NewBaseStage(name, &stage.BaseStageOptions{}), deps: deps}
}

func (s *contentDependenciesStub) GetContentDependencies(_ context.Context, _ stage.Conveyor, _ container_backend.BuildContextArchiver) (string, error) {
	return s.deps, s.err
}

var _ storage.PrimaryStagesStorage = (*anchorPrimaryStagesStorage)(nil)

type anchorPrimaryStagesStorage struct {
	storage.PrimaryStagesStorage
}

func (m *anchorLookupStorageManager) GetStagesStorage() storage.PrimaryStagesStorage {
	return m.primaryStagesStorage
}

func (m *anchorLookupStorageManager) GetStageDescSetByDigestFromStagesStorageCached(_ context.Context, _, _ string, _ int64, stagesStorage storage.StagesStorage) (imagePkg.StageDescSet, error) {
	if stagesStorage == m.secondaryStagesStorage {
		m.cachedSecondaryLookups++
		return m.inSecondary, nil
	}

	m.cachedPrimaryLookups++
	return m.inPrimary, nil
}
