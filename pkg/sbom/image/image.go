package image

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/distribution/reference"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/extract"
)

const (
	scratchImageName = "scratch"
	TagSuffix        = "-sbom"
)

func IsScratchRef(imageRef string) bool {
	if imageRef == "" {
		return false
	}

	if imageRef == scratchImageName {
		return true
	}

	ref, err := reference.ParseAnyReference(imageRef)
	if err != nil {
		return false
	}

	named, ok := ref.(reference.Named)
	if !ok {
		return false
	}

	path := reference.Path(named)

	return path == scratchImageName || strings.HasSuffix(path, "/"+scratchImageName)
}

func ImageName(name string) string {
	return fmt.Sprintf("%s%s", name, TagSuffix)
}

func BaseImageName(repo, tag string) string {
	return ImageName(fmt.Sprintf("%s:%s", repo, tag))
}

func PullCycloneDX16BOM(ctx context.Context, registry docker_registry.Interface, reference string) (*cdx.BOM, error) {
	var buf bytes.Buffer
	if err := registry.PullImageArchive(ctx, &buf, reference); err != nil {
		return nil, fmt.Errorf("pull image archive: %w", err)
	}

	archiveBytes := buf.Bytes()
	opener := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(archiveBytes)), nil
	}

	sbomJSON, err := extract.FromImageBytes(opener)
	if err != nil {
		return nil, fmt.Errorf("extract SBOM from image: %w", err)
	}

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(sbomJSON)
	if err != nil {
		return nil, fmt.Errorf("parse CycloneDX BOM: %w", err)
	}

	return bom, nil
}
