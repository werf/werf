package image

import (
	"context"
	"fmt"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/distribution/reference"

	"github.com/werf/werf/v2/pkg/oci/artifact"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
)

const (
	scratchImageName     = "scratch"
	DSSEMediaType        = "application/vnd.dsse.envelope.v1+json"
	InTotoMediaType      = "application/vnd.in-toto+json"
	CycloneDX16Predicate = "https://cyclonedx.org/bom/v1.6"
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

func FallbackTag(parentDigest string) string {
	return artifact.FallbackTag(parentDigest)
}

func PushSBOM(ctx context.Context, bomJSON []byte, repo, parentDigest, imageName, checksum, targetPlatform string) error {
	digestHex, err := artifact.DigestHex(parentDigest)
	if err != nil {
		return fmt.Errorf("extract digest hex: %w", err)
	}

	stmtBytes, err := WrapInInTotoStatement(bomJSON, CycloneDX16Predicate, repo, digestHex)
	if err != nil {
		return fmt.Errorf("wrap BOM in in-toto statement: %w", err)
	}

	envelopeBytes, err := WrapInDSSE(stmtBytes, InTotoMediaType)
	if err != nil {
		return fmt.Errorf("wrap in-toto statement in DSSE: %w", err)
	}

	store := artifact.NewOCIStore(repo, imageName)
	return store.Attach(ctx, parentDigest, DSSEMediaType, envelopeBytes, checksum, targetPlatform)
}

func PullSBOM(ctx context.Context, repo, parentDigest, imageName string) ([]byte, error) {
	store := artifact.NewOCIStore(repo, imageName)

	var envelopeJSON []byte
	if imageName != "" {
		var err error
		envelopeJSON, err = store.GetAttachedContent(ctx, parentDigest, DSSEMediaType)
		if err != nil {
			return nil, fmt.Errorf("get attached SBOM: %w", err)
		}
	} else {
		var err error
		envelopeJSON, err = store.GetAttachedContentAny(ctx, parentDigest, DSSEMediaType)
		if err != nil {
			return nil, fmt.Errorf("get attached SBOM: %w", err)
		}
	}

	stmtBytes, err := UnwrapDSSE(envelopeJSON, InTotoMediaType)
	if err != nil {
		return nil, fmt.Errorf("unwrap DSSE envelope: %w", err)
	}

	predicate, predicateType, err := UnwrapInTotoStatement(stmtBytes)
	if err != nil {
		return nil, fmt.Errorf("unwrap in-toto statement: %w", err)
	}

	if predicateType != CycloneDX16Predicate {
		return nil, fmt.Errorf("unexpected in-toto predicate type %q, expected %q", predicateType, CycloneDX16Predicate)
	}

	return []byte(predicate), nil
}

// PullSBOMByTag resolves the image tag to a digest and returns the attached SBOM.
func PullSBOMByTag(ctx context.Context, repo, tag, imageName string) ([]byte, error) {
	parentDigest, err := artifact.ResolveTag(ctx, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("resolve image tag: %w", err)
	}

	return PullSBOM(ctx, repo, parentDigest, imageName)
}

func PullCycloneDX16BOM(ctx context.Context, repo, parentDigest, imageName string) (*cdx.BOM, error) {
	bomJSON, err := PullSBOM(ctx, repo, parentDigest, imageName)
	if err != nil {
		return nil, err
	}

	bom, err := cyclonedxutil.BuildCycloneDX16BOMFromJSON(bomJSON)
	if err != nil {
		return nil, fmt.Errorf("parse CycloneDX BOM: %w", err)
	}

	return bom, nil
}

func PullCycloneDX16BOMContent(ctx context.Context, envelopeJSON []byte) ([]byte, error) {
	stmtBytes, err := UnwrapDSSE(envelopeJSON, InTotoMediaType)
	if err != nil {
		return nil, err
	}

	predicate, predicateType, err := UnwrapInTotoStatement(stmtBytes)
	if err != nil {
		return nil, err
	}

	if predicateType != CycloneDX16Predicate {
		return nil, fmt.Errorf("unexpected in-toto predicate type %q, expected %q", predicateType, CycloneDX16Predicate)
	}

	return []byte(predicate), nil
}
