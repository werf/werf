package gost

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gost SBOM validator", func() {
	DescribeTable("Validate",
		func(bom *cdx.BOM, expectedErrMatcher OmegaMatcher) {
			err := Validate(bom)
			Expect(err).To(expectedErrMatcher)
		},
		Entry("should fail if BOM is nil",
			nil,
			MatchError("BOM is required")),
		Entry("should fail if SpecVersion is not 1.6",
			&cdx.BOM{SpecVersion: cdx.SpecVersion1_5},
			MatchError(ContainSubstring("requires CycloneDX version 1.6"))),
		Entry("should fail if GOST properties are missing in metadata component",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Metadata: &cdx.Metadata{
					Component: &cdx.Component{Name: "test"},
				},
			},
			MatchError(ContainSubstring("missing mandatory GOST properties"))),
		Entry("should fail if GOST properties are missing in components",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Components: &[]cdx.Component{
					{Name: "test-comp"},
				},
			},
			MatchError(ContainSubstring("missing mandatory GOST properties"))),
		Entry("should fail if GOST properties have invalid values",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Components: &[]cdx.Component{
					{
						Name: "test-comp",
						Properties: &[]cdx.Property{
							{Name: PropertyAttackSurface, Value: "invalid"},
							{Name: PropertySecurityFunction, Value: "yes"},
						},
					},
				},
			},
			MatchError(ContainSubstring("invalid value for GOST:attack_surface"))),
		Entry("should succeed if GOST properties are present and valid (yes/no)",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Components: &[]cdx.Component{
					{
						Name: "test-comp",
						Properties: &[]cdx.Property{
							{Name: PropertyAttackSurface, Value: "yes"},
							{Name: PropertySecurityFunction, Value: "no"},
						},
					},
				},
			},
			Succeed()),
		Entry("should succeed if GOST properties are present and valid (inherit)",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Components: &[]cdx.Component{
					{
						Name: "test-comp",
						Properties: &[]cdx.Property{
							{Name: PropertyAttackSurface, Value: "inherit"},
							{Name: PropertySecurityFunction, Value: "inherit"},
						},
					},
				},
			},
			Succeed()),
	)
})
