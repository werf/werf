package finder

import (
	"archive/tar"
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/container_backend/stream_reader"
	"github.com/werf/werf/v2/pkg/sbom"
)

type finder struct {
	backend container_backend.ContainerBackend
}

func NewFinder(backend container_backend.ContainerBackend) *finder {
	return &finder{
		backend: backend,
	}
}

// FindArtifactFile finds SBOM artifact file in built images
func (f *finder) FindArtifactFile(ctx context.Context, images []*image.Image, imgNameToFind string) (*stream_reader.File, error) {
	foundImage, ok := lo.Find(images, func(item *image.Image) bool {
		return item.Name == imgNameToFind
	})
	if !ok {
		return nil, fmt.Errorf("unable to find requested image %q", imgNameToFind)
	}

	sbomImageName := sbom.ImageName(foundImage.GetLastNonEmptyStage().GetStageImage().Image.GetStageDesc().Info.Name)

	bReader, err := f.backend.DumpImage(ctx, sbomImageName)
	if err != nil {
		return nil, fmt.Errorf("unable to stream image %q: %w", sbomImageName, err)
	}
	fsStreamReader, err := stream_reader.NewFileSystemStreamReader(tar.NewReader(bReader))
	if err != nil {
		return nil, fmt.Errorf("unable to create file system stream reader: %w", err)
	}

	artifactFile, ok, err := fsStreamReader.Find(func(file *stream_reader.File) bool {
		// TODO: assume we have only one artifact file
		return !file.Info().IsDir()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to search artifact file in SBOM image %q: %w", sbomImageName, err)
	}
	if !ok {
		return nil, fmt.Errorf("artifact file is not found in SBOM image %q", sbomImageName)
	}

	return artifactFile, nil
}
