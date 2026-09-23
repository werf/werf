package build

import (
	"context"

	"github.com/onsi/gomega"

	buildImage "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/build/stage"
	"github.com/werf/werf/v2/pkg/container_backend"
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
