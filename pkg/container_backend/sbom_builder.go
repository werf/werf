package container_backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/werf/v2/pkg/sbom/scanner"
)

type SBOMImageBuilder struct {
	Backend ContainerBackend
}

func NewSBOMImageBuilder(backend ContainerBackend) *SBOMImageBuilder {
	return &SBOMImageBuilder{Backend: backend}
}

func (b *SBOMImageBuilder) BuildImage(ctx context.Context, source SBOMSource, scanOpts scanner.ScanOptions, labels []string) (string, error) {
	wt := scanner.NewWorkingTree()

	billNames := scanner.BillNamesFromCommands(scanOpts.Commands)
	if err := wt.Create(ctx, os.TempDir(), billNames); err != nil {
		return "", fmt.Errorf("create working tree: %w", err)
	}
	defer wt.Cleanup(ctx)

	if err := source.Fill(ctx, wt); err != nil {
		return "", err
	}

	billNamesFromScanner := scanner.BillNamesFromCommands(scanOpts.Commands)
	contextAddFiles := make([]string, 0, len(billNamesFromScanner)+1)
	for _, billName := range billNamesFromScanner {
		contextAddFiles = append(contextAddFiles, filepath.Join(wt.BillsDir(), billName))
	}
	contextAddFiles = append(contextAddFiles, wt.Containerfile())

	archive := NewSbomContextArchiver(wt.RootDir())
	if err := archive.Create(ctx, BuildContextArchiveCreateOptions{
		DockerfileRelToContextPath: wt.Containerfile(),
		ContextAddFiles:            contextAddFiles,
	}); err != nil {
		return "", fmt.Errorf("unable to create build context archive: %w", err)
	}

	imgId, err := b.Backend.BuildDockerfile(ctx, wt.ContainerfileContent(), BuildDockerfileOpts{
		DockerfileCtxRelPath: wt.Containerfile(),
		BuildContextArchive:  archive,
		Labels:               labels,
		Quiet:                true,
	})
	if err != nil {
		return "", fmt.Errorf("unable to build SBOM image: %w", err)
	}

	return imgId, nil
}
