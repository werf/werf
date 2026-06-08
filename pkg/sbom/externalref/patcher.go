package externalref

import (
	"context"
	"fmt"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

const EnvName = "WERF_EXTERNAL_REFS_SERVER_URL"

type ExternalRefPatcher struct {
	enricher *Enricher
}

func NewExternalRefPatcher() (*ExternalRefPatcher, error) {
	serverURL := os.Getenv(EnvName)
	if serverURL == "" {
		return nil, fmt.Errorf("%s env var is required", EnvName)
	}

	svc := NewService(ServiceConfig{ServerURL: serverURL})
	return &ExternalRefPatcher{enricher: NewEnricher(svc.Resolve)}, nil
}

func (p *ExternalRefPatcher) Apply(ctx context.Context, bom *cdx.BOM) (*cdx.BOM, error) {
	if err := p.enricher.Enrich(ctx, bom); err != nil {
		return bom, fmt.Errorf("enrich external references: %w", err)
	}

	return bom, nil
}
