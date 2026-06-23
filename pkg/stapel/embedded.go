package stapel

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/containerd/containerd/platforms"

	"github.com/werf/werf/v2/pkg/container_backend/thirdparty/platformutil"
	"github.com/werf/werf/v2/pkg/docker"
)

//go:embed embed/linux/amd64/werf-stapel-toolchain.tar.gz
var embeddedLinuxAmd64 []byte

//go:embed embed/linux/amd64/werf-stapel-toolchain.tar.sha256
var embeddedLinuxAmd64Sha256 string

//go:embed embed/linux/arm64/werf-stapel-toolchain.tar.gz
var embeddedLinuxArm64 []byte

//go:embed embed/linux/arm64/werf-stapel-toolchain.tar.sha256
var embeddedLinuxArm64Sha256 string

type embeddedImage struct {
	gzipData       []byte
	expectedSha256 string
}

func embeddedImageForPlatform(targetPlatform string) (embeddedImage, bool) {
	normalized, err := normalizeEmbeddedPlatform(targetPlatform)
	if err != nil {
		return embeddedImage{}, false
	}

	switch normalized {
	case "linux/amd64":
		return embeddedImage{gzipData: embeddedLinuxAmd64, expectedSha256: strings.TrimSpace(embeddedLinuxAmd64Sha256)}, true
	case "linux/arm64":
		return embeddedImage{gzipData: embeddedLinuxArm64, expectedSha256: strings.TrimSpace(embeddedLinuxArm64Sha256)}, true
	default:
		return embeddedImage{}, false
	}
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

func loadEmbeddedImage(ctx context.Context, targetPlatform string) error {
	img, ok := embeddedImageForPlatform(targetPlatform)
	if !ok {
		return fmt.Errorf("no embedded stapel image for platform %q", targetPlatform)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(img.gzipData))
	if err != nil {
		return fmt.Errorf("init gzip reader for embedded stapel image: %w", err)
	}
	defer gzReader.Close()

	tarData, err := io.ReadAll(gzReader)
	if err != nil {
		return fmt.Errorf("decompress embedded stapel image: %w", err)
	}

	sum := sha256.Sum256(tarData)
	actualSha256 := hex.EncodeToString(sum[:])
	if actualSha256 != img.expectedSha256 {
		return fmt.Errorf("embedded stapel image integrity check failed: expected sha256 %s, got %s", img.expectedSha256, actualSha256)
	}

	if _, err := docker.CliLoadFromStream(ctx, bytes.NewReader(tarData)); err != nil {
		return fmt.Errorf("load embedded stapel image: %w", err)
	}

	return nil
}
