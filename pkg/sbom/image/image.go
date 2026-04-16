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

func PullRawSbom(ctx context.Context, registry docker_registry.Interface, reference string) ([]byte, error) {
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

	return sbomJSON, nil
}

func PullCycloneDX16BOM(ctx context.Context, registry docker_registry.Interface, reference string) (*cdx.BOM, error) {
	sbomJSON, err := PullRawSbom(ctx, registry, reference)
	if err != nil {
		return nil, err
	}

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(sbomJSON)
	if err != nil {
		return nil, fmt.Errorf("parse CycloneDX BOM: %w", err)
	}

	return bom, nil
}

func BuildDigestToTagIndex(ctx context.Context, registry docker_registry.Interface, repo string) (map[string]string, error) {
	tags, err := registry.Tags(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("list tags for %s: %w", repo, err)
	}

	result := make(map[string]string, len(tags))
	for _, tag := range tags {
		if strings.HasSuffix(tag, TagSuffix) {
			continue
		}

		ref := fmt.Sprintf("%s:%s", repo, tag)

		info, err := registry.TryGetRepoImage(ctx, ref)
		if err != nil {
			return nil, fmt.Errorf("get image info for %s: %w", ref, err)
		}
		if info == nil {
			continue
		}

		d := info.GetDigest()
		if d == "" {
			continue
		}

		if _, exists := result[d]; !exists {
			result[d] = tag
		}
	}

	return result, nil
}

func ResolveSBOMReference(repo, imageDigest string, digestToTag map[string]string) (string, error) {
	tag, ok := digestToTag[imageDigest]
	if !ok {
		return "", fmt.Errorf("no tag found for digest %s in repo %s", imageDigest, repo)
	}

	return BaseImageName(repo, tag), nil
}
