//go:build embedstapel

package stapel

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"

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

func loadEmbeddedImage(ctx context.Context, targetPlatform string) error {
	img, ok := embeddedImageForPlatform(targetPlatform)
	if !ok {
		return fmt.Errorf("no embedded stapel image for platform %q", targetPlatform)
	}

	tarData, err := decompressAndVerify(img.gzipData, img.expectedSha256)
	if err != nil {
		return err
	}

	if _, err := docker.CliLoadFromStream(ctx, bytes.NewReader(tarData)); err != nil {
		return fmt.Errorf("load embedded stapel image: %w", err)
	}

	return nil
}
