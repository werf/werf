package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

var platforms = []v1.Platform{
	{OS: "linux", Architecture: "amd64"},
	{OS: "linux", Architecture: "arm64"},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "embed-stapel: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	imageRef := os.Getenv("WERF_STAPEL_IMAGE_NAME")
	if imageRef == "" {
		imageRef = "registry.werf.io/werf/stapel"
	}

	version := os.Getenv("WERF_STAPEL_IMAGE_VERSION")
	if version == "" {
		return fmt.Errorf("WERF_STAPEL_IMAGE_VERSION is required")
	}

	embedRoot := os.Getenv("STAPEL_EMBED_ROOT")
	if embedRoot == "" {
		embedRoot = filepath.Join("pkg", "stapel", "embed")
	}

	taggedRef := fmt.Sprintf("%s:%s", imageRef, version)

	tag, err := name.NewTag(taggedRef)
	if err != nil {
		return fmt.Errorf("parse image tag %q: %w", taggedRef, err)
	}

	for _, platform := range platforms {
		if err := buildPlatform(tag, platform, embedRoot); err != nil {
			return fmt.Errorf("build %s/%s: %w", platform.OS, platform.Architecture, err)
		}
	}

	return nil
}

func buildPlatform(tag name.Tag, platform v1.Platform, embedRoot string) error {
	platformDir := filepath.Join(embedRoot, platform.OS, platform.Architecture)
	if err := os.MkdirAll(platformDir, 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", platformDir, err)
	}

	img, err := crane.Pull(tag.String(), crane.WithPlatform(&platform))
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}

	tarPath := filepath.Join(platformDir, "werf-stapel-toolchain.tar.gz")
	sha256Path := filepath.Join(platformDir, "werf-stapel-toolchain.tar.sha256")

	tarTmpFile, err := os.CreateTemp(platformDir, "werf-stapel-toolchain.tar.gz.*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", platformDir, err)
	}
	tarTmpPath := tarTmpFile.Name()
	defer os.Remove(tarTmpPath)

	hasher := sha256.New()
	gzWriter := gzip.NewWriter(tarTmpFile)

	if err := tarball.Write(tag, img, io.MultiWriter(gzWriter, hasher)); err != nil {
		tarTmpFile.Close()
		return fmt.Errorf("write tarball: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		tarTmpFile.Close()
		return fmt.Errorf("close gzip writer: %w", err)
	}
	if err := tarTmpFile.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tarTmpPath, err)
	}

	decompressedSha256 := hex.EncodeToString(hasher.Sum(nil))

	sha256TmpPath := sha256Path + ".tmp"
	if err := os.WriteFile(sha256TmpPath, []byte(decompressedSha256+"\n"), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", sha256TmpPath, err)
	}
	defer os.Remove(sha256TmpPath)

	if err := os.Rename(tarTmpPath, tarPath); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tarTmpPath, tarPath, err)
	}
	if err := os.Rename(sha256TmpPath, sha256Path); err != nil {
		return fmt.Errorf("rename %s to %s: %w", sha256TmpPath, sha256Path, err)
	}

	fmt.Printf("Embedded stapel for %s/%s (sha256 %s)\n", platform.OS, platform.Architecture, decompressedSha256)
	return nil
}
