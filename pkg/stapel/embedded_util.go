package stapel

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/containerd/containerd/platforms"

	"github.com/werf/werf/v2/pkg/container_backend/thirdparty/platformutil"
	"github.com/werf/werf/v2/pkg/docker"
)

type embeddedImage struct {
	gzipData       []byte
	expectedSha256 string
}

func normalizeEmbeddedPlatform(targetPlatform string) (string, error) {
	if targetPlatform == "" {
		targetPlatform = docker.GetDefaultPlatform()
	}

	spec, err := platformutil.ParsePlatform(targetPlatform)
	if err != nil {
		return "", fmt.Errorf("parse target platform %q: %w", targetPlatform, err)
	}

	return platforms.Format(platforms.Normalize(spec)), nil
}

func decompressAndVerify(gzipData []byte, expectedSha256 string) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(gzipData))
	if err != nil {
		return nil, fmt.Errorf("init gzip reader for embedded stapel image: %w", err)
	}
	defer gzReader.Close()

	tarData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("decompress embedded stapel image: %w", err)
	}

	sum := sha256.Sum256(tarData)
	actualSha256 := hex.EncodeToString(sum[:])
	if actualSha256 != expectedSha256 {
		return nil, fmt.Errorf("embedded stapel image integrity check failed: expected sha256 %s, got %s", expectedSha256, actualSha256)
	}

	return tarData, nil
}
