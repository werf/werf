package artifact

import (
	"context"
	"fmt"

	"github.com/opencontainers/go-digest"

	"github.com/werf/werf/v2/pkg/docker_registry"
)

// DigestHex returns the hex-encoded portion of a digest string (e.g. "sha256:abc123..." → "abc123...").
// It uses the go-digest parser to correctly handle any hash algorithm, not just SHA-256.
func DigestHex(parentDigest string) (string, error) {
	d, err := digest.Parse(parentDigest)
	if err != nil {
		return "", fmt.Errorf("parse digest %q: %w", parentDigest, err)
	}
	return d.Encoded(), nil
}

// ResolveTag resolves an image tag to its digest by querying the container registry
// using the global docker_registry API (which handles auth, TLS, etc.).
func ResolveTag(ctx context.Context, repo, tag string) (string, error) {
	ref := fmt.Sprintf("%s:%s", repo, tag)
	desc, err := docker_registry.API().GetRepoImageDesc(ctx, ref)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", ref, err)
	}
	return desc.Digest.String(), nil
}
