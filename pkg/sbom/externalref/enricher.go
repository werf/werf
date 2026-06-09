package externalref

import (
	"context"
	"fmt"
	"sync"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"golang.org/x/sync/errgroup"

	"github.com/werf/logboek"
)

func validateRefKind(kind string) error {
	switch cdx.ExternalReferenceType(kind) {
	case cdx.ERTypeVCS, cdx.ERTypeWebsite, cdx.ERTypeIssueTracker, cdx.ERTypeAdvisories,
		cdx.ERTypeBOM, cdx.ERTypeChat, cdx.ERTypeDocumentation, cdx.ERTypeDistribution,
		cdx.ERTypeLicense, cdx.ERTypeOther, cdx.ERTypeReleaseNotes, cdx.ERTypeSecurityContact,
		cdx.ERTypeSocial, cdx.ERTypeSupport, cdx.ERTypeEvidence, cdx.ERTypeFormulation,
		cdx.ERTypeConfiguration, cdx.ERTypeBuildMeta, cdx.ERTypeBuildSystem,
		cdx.ERTypeAttestation, cdx.ERTypeThreatModel, cdx.ERTypeRiskAssessment,
		cdx.ERTypeMaturityReport, cdx.ERTypeComponentAnalysisReport, cdx.ERTypeDynamicAnalysisReport,
		cdx.ERTypeStaticAnalysisReport, cdx.ERTypePentestReport, cdx.ERTypeCertificationReport,
		cdx.ERTypeQualityMetrics, cdx.ERTypePOAM, cdx.ERTypeRuntimeAnalysisReport,
		cdx.ERTypeExploitabilityStatement, cdx.ERTypeAdversaryModel, cdx.ERTypeModelCard,
		cdx.ERTypeDistributionIntake, cdx.ERTypeDigitalSignature, cdx.ERTypeElectronicSignature,
		cdx.ERTypeCodifiedInfrastructure, cdx.ERTypeLog, cdx.ERTypeMailingList,
		cdx.ERTypeRFC9116, cdx.ERTypeSourceDistribution, cdx.ERTypeVulnerabilityAssertion:
		return nil
	default:
		return fmt.Errorf("enrich: unknown external reference kind %q", kind)
	}
}

func purlNotExpected(ct cdx.ComponentType) bool {
	switch ct {
	case cdx.ComponentTypeOS,
		cdx.ComponentTypeDevice,
		cdx.ComponentTypeDeviceDriver,
		cdx.ComponentTypeFile,
		cdx.ComponentTypeFirmware,
		cdx.ComponentTypePlatform,
		cdx.ComponentTypeData,
		cdx.ComponentTypeMachineLearningModel,
		cdx.ComponentTypeCryptographicAsset:
		return true
	default:
		return false
	}
}

type Enricher struct {
	resolve func(ctx context.Context, purl string) (*ResolveResult, error)
}

func NewEnricher(resolve func(ctx context.Context, purl string) (*ResolveResult, error)) *Enricher {
	return &Enricher{resolve: resolve}
}

func (e *Enricher) Enrich(ctx context.Context, bom *cdx.BOM) error {
	if bom == nil || bom.Components == nil {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	var seen sync.Map

	components := *bom.Components
	for i := range components {
		comp := &components[i]
		g.Go(func() error {
			if comp.PackageURL == "" {
				if purlNotExpected(comp.Type) {
					return nil
				}
				return fmt.Errorf("enrich: component %q (type %q) has no purl", comp.Name, comp.Type)
			}

			res, err := e.resolve(ctx, comp.PackageURL)
			if err != nil {
				return fmt.Errorf("enrich %q: %w", comp.PackageURL, err)
			}

			if err := validateRefKind(res.Kind); err != nil {
				return fmt.Errorf("enrich %q: %w", comp.PackageURL, err)
			}

			extRef := cdx.ExternalReference{
				URL:  res.URL,
				Type: cdx.ExternalReferenceType(res.Kind),
			}

			if comp.ExternalReferences == nil {
				comp.ExternalReferences = &[]cdx.ExternalReference{}
			}
			*comp.ExternalReferences = append(*comp.ExternalReferences, extRef)

			seen.Store(res.URL+"|"+res.Kind, extRef)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	var bomRefs []cdx.ExternalReference
	seen.Range(func(_, value any) bool {
		bomRefs = append(bomRefs, value.(cdx.ExternalReference))
		return true
	})

	if len(bomRefs) > 0 {
		bom.ExternalReferences = &bomRefs
		logboek.Context(ctx).Debug().LogF("Enriched SBOM with %d external references\n", len(bomRefs))
	}

	return nil
}
