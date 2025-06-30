package stage

import (
	"context"
	"fmt"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"

	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/werf/exec"
)

const (
	annoNameBuildTimestamp   = "io.deckhouse.deliverykit.build-timestamp"
	annoNameDMVerityRootHash = "io.deckhouse.deliverykit.dm-verity-root-hash"

	mkfsBuildDate   = "2025-06-24T18:50:50Z"
	magicVeritySalt = "dc0f616e4bf75776061d5ffb7a6f45e1313b7cc86f3aa49b68de4f6d187bad2b" // sha256("Я ненавижу тупые приказы ФСТЭК")
)

type ManifestStage struct {
	*BaseStage
	imageSpec *config.ImageSpec
}

func GenerateManifestStage(baseStageOptions *BaseStageOptions) *ManifestStage {
	return newManifestStage(baseStageOptions)
}

func newManifestStage(baseStageOptions *BaseStageOptions) *ManifestStage {
	return &ManifestStage{
		BaseStage: NewBaseStage(Manifest, baseStageOptions),
	}
}

func (s *ManifestStage) IsBuildable() bool {
	return false
}

func (s *ManifestStage) IsMutable() bool {
	return true
}

func (s *ManifestStage) PrepareImage(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) error {
	return nil
}

func (s *ManifestStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	return "", nil
}

func (s *ManifestStage) MutateImage(ctx context.Context, registry docker_registry.Interface, prevBuiltImage, stageImage *StageImage) error {
	srcRef := prevBuiltImage.Image.Name()
	destRef := stageImage.Image.Name()

	return registry.MutateAndPushImage(
		ctx,
		srcRef,
		destRef,
		api.WithLayersMutation(processLayers),
	)
}

func processLayers(ctx context.Context, layers []v1.Layer) ([]mutate.Addendum, error) {
	var result []mutate.Addendum

	for _, layer := range layers {
		rc, err := layer.Uncompressed()
		if err != nil {
			return nil, fmt.Errorf("get uncompressed layer: %w", err)
		}
		defer rc.Close()

		tmpDir, err := os.MkdirTemp("", "layer-erofs")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(tmpDir)

		erofsPath := filepath.Join(tmpDir, "layer.erofs.img")
		hashPath := filepath.Join(tmpDir, "layer.hash.img")

		mkfsBuildTime, err := time.Parse(time.RFC3339, mkfsBuildDate)
		if err != nil {
			panic(err)
		}

		mkfsBuildTimestamp := strconv.FormatInt(mkfsBuildTime.Unix(), 10)

		// Create EROFS image from layer tar
		mkfs := exec.CommandContextCancellation(ctx, "mkfs.erofs", "-Uclear", "-T"+mkfsBuildTimestamp, "-x-1", "-Enoinline_data", "--tar=-", erofsPath)
		mkfs.Stderr = os.Stderr
		mkfs.Stdin = rc

		if err := mkfs.Run(); err != nil {
			return nil, fmt.Errorf("mkfs.erofs: %w", err)
		}

		// Create dummy hash image
		dd := exec.CommandContextCancellation(ctx, "dd", "if=/dev/zero", fmt.Sprintf("of=%s", hashPath), "bs=1M", "count=4")
		if err := dd.Run(); err != nil {
			return nil, fmt.Errorf("dd: %w", err)
		}

		// Run veritysetup format
		veritysetup := exec.CommandContextCancellation(
			ctx,
			"veritysetup", "format",
			"--data-block-size=4096",
			"--hash-block-size=4096",
			"--salt="+magicVeritySalt,
			erofsPath,
			hashPath,
		)
		out, err := veritysetup.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("veritysetup: %v\n%s", err, string(out))
		}

		// Extract root hash from output
		rootHash := extractRootHash(string(out))
		if rootHash == "" {
			return nil, fmt.Errorf("failed to extract root hash")
		}

		addendum := mutate.Addendum{Layer: layer}
		addendum.Annotations = map[string]string{
			annoNameBuildTimestamp:   mkfsBuildTimestamp,
			annoNameDMVerityRootHash: rootHash,
		}
		result = append(result, addendum)
	}

	return result, nil
}

func extractRootHash(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "Root hash:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Root hash:"))
		}
	}
	return ""
}
