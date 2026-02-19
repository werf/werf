package container_backend

import (
	"context"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil"
	"github.com/werf/werf/v2/pkg/sbom/scanner"
)

type SBOMSource interface {
	Fill(ctx context.Context, wt *scanner.WorkingTree) error
}

type StaticSource struct {
	BOM *cdx.BOM
}

func NewStaticSource(bom *cdx.BOM) *StaticSource {
	return &StaticSource{BOM: bom}
}

func (s *StaticSource) Fill(ctx context.Context, wt *scanner.WorkingTree) error {
	bomJSON, err := cyclonedxutil.ToJSON(s.BOM)
	if err != nil {
		return fmt.Errorf("serialize BOM: %w", err)
	}

	if err := wt.WriteBOMToFirstFile(bomJSON); err != nil {
		return fmt.Errorf("write static BOM to working tree: %w", err)
	}

	return nil
}

type ScannerRunner func(ctx context.Context, wt *scanner.WorkingTree) error

type ScannerSource struct {
	runner ScannerRunner
}

func NewScannerSource(runner ScannerRunner) *ScannerSource {
	return &ScannerSource{runner: runner}
}

func (s *ScannerSource) Fill(ctx context.Context, wt *scanner.WorkingTree) error {
	if s.runner == nil {
		return fmt.Errorf("scanner runner is not provided")
	}
	return s.runner(ctx, wt)
}
