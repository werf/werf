package cyclonedxutil

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

func componentNames(bom *cdx.BOM) []string {
	if bom == nil || bom.Components == nil {
		return nil
	}
	comps := *bom.Components
	out := make([]string, 0, len(comps))
	for _, c := range comps {
		out = append(out, c.Name)
	}
	return out
}

func dependencyRefs(bom *cdx.BOM) []string {
	if bom == nil || bom.Dependencies == nil {
		return nil
	}
	deps := *bom.Dependencies
	out := make([]string, 0, len(deps))
	for _, d := range deps {
		out = append(out, d.Ref)
	}
	return out
}

var _ = Describe("MergeBOMs", func() {
	It("concatenates components in order", func() {
		baseBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components: &[]cdx.Component{
				{Name: "base-comp-1", Version: "1.0.0"},
				{Name: "base-comp-2", Version: "2.0.0"},
			},
		}
		importBOM1 := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components: &[]cdx.Component{
				{Name: "import1-comp", Version: "1.0.0"},
			},
		}
		importBOM2 := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components: &[]cdx.Component{
				{Name: "import2-comp", Version: "1.0.0"},
			},
		}
		fragmentBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components: &[]cdx.Component{
				{Name: "fragment-comp", Version: "1.0.0"},
			},
		}
		targetBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components: &[]cdx.Component{
				{Name: "target-comp", Version: "1.0.0"},
			},
		}

		result, err := MergeBOMs(targetBOM, MergeOpts{
			BaseBOM:     baseBOM,
			ImportBOMs:  []*cdx.BOM{importBOM1, importBOM2},
			FragmentBOM: fragmentBOM,
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(result.Components).ToNot(BeNil())
		Expect(componentNames(result)).To(Equal([]string{
			"base-comp-1",
			"base-comp-2",
			"import1-comp",
			"import2-comp",
			"fragment-comp",
			"target-comp",
		}))
	})

	It("takes metadata from target", func() {
		baseBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Metadata: &cdx.Metadata{
				Component: &cdx.Component{Name: "base-metadata-component"},
			},
			Components: &[]cdx.Component{},
		}
		targetBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Metadata: &cdx.Metadata{
				Component: &cdx.Component{Name: "target-metadata-component"},
			},
			Components: &[]cdx.Component{},
		}

		result, err := MergeBOMs(targetBOM, MergeOpts{BaseBOM: baseBOM})
		Expect(err).ToNot(HaveOccurred())

		Expect(result.Metadata).ToNot(BeNil())
		Expect(result.Metadata.Component).ToNot(BeNil())
		Expect(result.Metadata.Component.Name).To(Equal("target-metadata-component"))
	})

	It("sets correct BOM fields", func() {
		targetBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{}}

		result, err := MergeBOMs(targetBOM, MergeOpts{})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.BOMFormat).To(Equal(cdx.BOMFormat))
		Expect(result.SpecVersion).To(Equal(cdx.SpecVersion1_6))
		Expect(result.Version).To(Equal(1))
		Expect(result.SerialNumber).To(HavePrefix("urn:uuid:"))
		Expect(result.Dependencies).To(BeNil(), "no dependencies were provided, so result should have none")
		Expect(result.Declarations).To(BeNil(), "no declarations were provided, so result should have none")
	})

	It("generates new serial number", func() {
		targetBOM := &cdx.BOM{
			SpecVersion:  cdx.SpecVersion1_6,
			SerialNumber: "urn:uuid:old-serial-number",
			Components:   &[]cdx.Component{},
		}

		result, err := MergeBOMs(targetBOM, MergeOpts{})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.SerialNumber).To(HavePrefix("urn:uuid:"))
		Expect(result.SerialNumber).ToNot(Equal(targetBOM.SerialNumber))
	})

	It("handles nil target", func() {
		baseBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components: &[]cdx.Component{
				{Name: "base-comp", Version: "1.0.0"},
			},
		}

		result, err := MergeBOMs(nil, MergeOpts{BaseBOM: baseBOM})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Components).ToNot(BeNil())
		Expect(*result.Components).To(HaveLen(1))
		Expect(result.Metadata).To(BeNil())
	})

	It("deduplicates identical components from different BOMs", func() {
		duplicateComp := cdx.Component{Name: "duplicate-comp", Version: "1.0.0"}
		baseBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{duplicateComp}}
		targetBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{duplicateComp}}

		result, err := MergeBOMs(targetBOM, MergeOpts{BaseBOM: baseBOM})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Components).ToNot(BeNil())
		Expect(*result.Components).To(HaveLen(1))
		Expect((*result.Components)[0].Name).To(Equal("duplicate-comp"))
	})

	It("keeps components that differ in any field", func() {
		comp1 := cdx.Component{Name: "comp", Version: "1.0.0"}
		comp2 := cdx.Component{Name: "comp", Version: "2.0.0"}
		baseBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{comp1}}
		targetBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{comp2}}

		result, err := MergeBOMs(targetBOM, MergeOpts{BaseBOM: baseBOM})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Components).ToNot(BeNil())
		Expect(*result.Components).To(HaveLen(2))
	})

	It("returns error for unsupported spec version in target BOM", func() {
		targetBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_5,
			Components:  &[]cdx.Component{},
		}

		_, err := MergeBOMs(targetBOM, MergeOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported CycloneDX spec version"))
		Expect(err.Error()).To(ContainSubstring("1.5"))
	})

	It("returns error for unsupported spec version in base BOM", func() {
		baseBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion(100),
			Components:  &[]cdx.Component{},
		}
		targetBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{}}

		_, err := MergeBOMs(targetBOM, MergeOpts{BaseBOM: baseBOM})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported CycloneDX spec version"))
		Expect(err.Error()).To(ContainSubstring("SpecVersion(100)"))
	})

	It("returns error for unsupported spec version in import BOM", func() {
		importBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_5,
			Components:  &[]cdx.Component{},
		}
		targetBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{}}

		_, err := MergeBOMs(targetBOM, MergeOpts{ImportBOMs: []*cdx.BOM{importBOM}})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported CycloneDX spec version"))
	})

	It("succeeds when BOMs have matching 1.6 spec version", func() {
		baseBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components:  &[]cdx.Component{{Name: "base-comp"}},
		}
		targetBOM := &cdx.BOM{
			SpecVersion: cdx.SpecVersion1_6,
			Components:  &[]cdx.Component{{Name: "target-comp"}},
		}

		result, err := MergeBOMs(targetBOM, MergeOpts{BaseBOM: baseBOM})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Components).ToNot(BeNil())
		Expect(*result.Components).To(HaveLen(2))
	})

	DescribeTable("merges dependencies",
		func(target *cdx.BOM, opts MergeOpts, assert func(*cdx.BOM)) {
			result, err := MergeBOMs(target, opts)
			Expect(err).ToNot(HaveOccurred())
			assert(result)
		},

		Entry("concatenates in merge order (base → imports → fragment → target)",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Dependencies: &[]cdx.Dependency{
					{Ref: "target-ref", Dependencies: &[]string{"dep-e"}},
				},
			},
			MergeOpts{
				BaseBOM: &cdx.BOM{
					SpecVersion: cdx.SpecVersion1_6,
					Dependencies: &[]cdx.Dependency{
						{Ref: "base-ref-1", Dependencies: &[]string{"dep-a"}},
						{Ref: "base-ref-2"},
					},
				},
				ImportBOMs: []*cdx.BOM{
					{SpecVersion: cdx.SpecVersion1_6, Dependencies: &[]cdx.Dependency{{Ref: "import1-ref"}}},
					{SpecVersion: cdx.SpecVersion1_6, Dependencies: &[]cdx.Dependency{{Ref: "import2-ref"}}},
				},
				FragmentBOM: &cdx.BOM{
					SpecVersion:  cdx.SpecVersion1_6,
					Dependencies: &[]cdx.Dependency{{Ref: "fragment-ref"}},
				},
			},
			func(result *cdx.BOM) {
				Expect(dependencyRefs(result)).To(Equal([]string{
					"base-ref-1", "base-ref-2",
					"import1-ref", "import2-ref",
					"fragment-ref",
					"target-ref",
				}))
			},
		),

		Entry("returns nil when no BOMs have dependencies",
			&cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{{Name: "comp"}}},
			MergeOpts{BaseBOM: &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{{Name: "comp"}}}},
			func(result *cdx.BOM) {
				Expect(result.Dependencies).To(BeNil())
			},
		),

		Entry("deduplicates identical dependencies",
			&cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Dependencies: &[]cdx.Dependency{{Ref: "dup", Dependencies: &[]string{"dep-a"}}}},
			MergeOpts{BaseBOM: &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Dependencies: &[]cdx.Dependency{{Ref: "dup", Dependencies: &[]string{"dep-a"}}}}},
			func(result *cdx.BOM) {
				Expect(result.Dependencies).ToNot(BeNil())
				Expect(*result.Dependencies).To(HaveLen(1))
				Expect((*result.Dependencies)[0].Ref).To(Equal("dup"))
			},
		),

		Entry("handles nil target",
			nil,
			MergeOpts{BaseBOM: &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Dependencies: &[]cdx.Dependency{{Ref: "base-ref", Dependencies: &[]string{"dep-a"}}}}},
			func(result *cdx.BOM) {
				Expect(result.Dependencies).ToNot(BeNil())
				Expect(*result.Dependencies).To(HaveLen(1))
				Expect((*result.Dependencies)[0].Ref).To(Equal("base-ref"))
			},
		),

		Entry("preserves dependsOn and provides fields",
			&cdx.BOM{SpecVersion: cdx.SpecVersion1_6},
			MergeOpts{BaseBOM: &cdx.BOM{
				SpecVersion:  cdx.SpecVersion1_6,
				Dependencies: &[]cdx.Dependency{{Ref: "ref-1", Dependencies: &[]string{"dep-a"}, Provides: &[]string{"prov-a"}}},
			}},
			func(result *cdx.BOM) {
				Expect(result.Dependencies).ToNot(BeNil())
				Expect(*result.Dependencies).To(HaveLen(1))
				dep := (*result.Dependencies)[0]
				Expect(dep.Ref).To(Equal("ref-1"))
				Expect(*dep.Dependencies).To(Equal([]string{"dep-a"}))
				Expect(*dep.Provides).To(Equal([]string{"prov-a"}))
			},
		),
	)

	DescribeTable("merges declarations",
		func(target *cdx.BOM, opts MergeOpts, assert func(*cdx.BOM)) {
			result, err := MergeBOMs(target, opts)
			Expect(err).ToNot(HaveOccurred())
			assert(result)
		},

		Entry("returns nil when no BOMs have declarations",
			&cdx.BOM{SpecVersion: cdx.SpecVersion1_6},
			MergeOpts{BaseBOM: &cdx.BOM{SpecVersion: cdx.SpecVersion1_6}},
			func(result *cdx.BOM) {
				Expect(result.Declarations).To(BeNil())
			},
		),

		Entry("concatenates assessors from multiple BOMs",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Assessors: &[]cdx.Assessor{{BOMRef: "target-assessor", ThirdParty: true}},
				},
			},
			MergeOpts{BaseBOM: &cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Assessors: &[]cdx.Assessor{{BOMRef: "base-assessor", ThirdParty: false}},
				},
			}},
			func(result *cdx.BOM) {
				Expect(result.Declarations).ToNot(BeNil())
				Expect(result.Declarations.Assessors).ToNot(BeNil())
				Expect(*result.Declarations.Assessors).To(HaveLen(2))
				Expect((*result.Declarations.Assessors)[0].BOMRef).To(Equal(cdx.BOMReference("base-assessor")))
				Expect((*result.Declarations.Assessors)[1].BOMRef).To(Equal(cdx.BOMReference("target-assessor")))
			},
		),

		Entry("concatenates claims and evidence",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Claims:   &[]cdx.Claim{{BOMRef: "target-claim", Predicate: "secure"}},
					Evidence: &[]cdx.DeclarationEvidence{{BOMRef: "target-evidence"}},
				},
			},
			MergeOpts{BaseBOM: &cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Claims:   &[]cdx.Claim{{BOMRef: "base-claim", Predicate: "compliant"}},
					Evidence: &[]cdx.DeclarationEvidence{{BOMRef: "base-evidence"}},
				},
			}},
			func(result *cdx.BOM) {
				Expect(result.Declarations).ToNot(BeNil())
				Expect(*result.Declarations.Claims).To(HaveLen(2))
				Expect(*result.Declarations.Evidence).To(HaveLen(2))
			},
		),

		Entry("concatenates targets (organizations, components, services)",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Targets: &cdx.Targets{
						Components: &[]cdx.Component{{Name: "target-comp"}},
						Services:   &[]cdx.Service{{Name: "target-svc"}},
					},
				},
			},
			MergeOpts{BaseBOM: &cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Targets: &cdx.Targets{
						Organizations: &[]cdx.OrganizationalEntity{{Name: "base-org"}},
						Components:    &[]cdx.Component{{Name: "base-comp"}},
					},
				},
			}},
			func(result *cdx.BOM) {
				Expect(result.Declarations).ToNot(BeNil())
				Expect(result.Declarations.Targets).ToNot(BeNil())
				Expect(*result.Declarations.Targets.Organizations).To(HaveLen(1))
				Expect(*result.Declarations.Targets.Components).To(HaveLen(2))
				Expect(*result.Declarations.Targets.Services).To(HaveLen(1))
			},
		),

		Entry("last affirmation wins",
			&cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Affirmation: &cdx.Affirmation{Statement: "target affirms"},
				},
			},
			MergeOpts{BaseBOM: &cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Affirmation: &cdx.Affirmation{Statement: "base affirms"},
				},
			}},
			func(result *cdx.BOM) {
				Expect(result.Declarations).ToNot(BeNil())
				Expect(result.Declarations.Affirmation).ToNot(BeNil())
				Expect(result.Declarations.Affirmation.Statement).To(Equal("target affirms"))
			},
		),

		Entry("handles nil target with declarations in base",
			nil,
			MergeOpts{BaseBOM: &cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Assessors: &[]cdx.Assessor{{BOMRef: "base-assessor"}},
				},
			}},
			func(result *cdx.BOM) {
				Expect(result.Declarations).ToNot(BeNil())
				Expect(*result.Declarations.Assessors).To(HaveLen(1))
			},
		),

		Entry("does not carry over declarations signature",
			&cdx.BOM{SpecVersion: cdx.SpecVersion1_6},
			MergeOpts{BaseBOM: &cdx.BOM{
				SpecVersion: cdx.SpecVersion1_6,
				Declarations: &cdx.Declarations{
					Assessors: &[]cdx.Assessor{{BOMRef: "assessor-1"}},
					Signature: &cdx.JSFSignature{},
				},
			}},
			func(result *cdx.BOM) {
				Expect(result.Declarations).ToNot(BeNil())
				Expect(result.Declarations.Signature).To(BeNil())
			},
		),
	)
})

var _ = Describe("ToJSON", func() {
	It("serializes BOM and contains required $schema", func() {
		bom := &cdx.BOM{
			BOMFormat:   cdx.BOMFormat,
			SpecVersion: cdx.SpecVersion1_6,
			Version:     1,
			Components: &[]cdx.Component{
				{Name: "test-comp", Version: "1.0.0"},
			},
		}

		data, err := ToJSON(bom)
		Expect(err).ToNot(HaveOccurred())
		Expect(data).ToNot(BeEmpty())

		// required by CycloneDX JSON format
		Expect(string(data)).To(ContainSubstring(`"$schema":"http://cyclonedx.org/schema/bom-1.6.schema.json"`))
	})
})

var _ = Describe("MergeOpts", func() {
	DescribeTable("IsEmpty",
		func(opts MergeOpts, expected bool) {
			Expect(opts.IsEmpty()).To(Equal(expected))
		},
		Entry("empty opts", MergeOpts{}, true),
		Entry("with base BOM", MergeOpts{BaseBOM: &cdx.BOM{}}, false),
		Entry("with import BOMs", MergeOpts{ImportBOMs: []*cdx.BOM{{}}}, false),
		Entry("with fragment BOM", MergeOpts{FragmentBOM: &cdx.BOM{}}, false),
		Entry("with empty import slice", MergeOpts{ImportBOMs: []*cdx.BOM{}}, true),
	)

	DescribeTable("Checksum",
		func(opts MergeOpts, expected string) {
			Expect(opts.Checksum()).To(Equal(expected))
		},
		Entry("empty opts", MergeOpts{}, ""),
		Entry("empty opts with GOST configuration (should be invariant)",
			MergeOpts{
				Gost: gost.Config{
					AttackSurface:    gost.GostValueYes,
					SecurityFunction: gost.GostValueInherit,
				},
			},
			""),
		Entry("opts with BaseBOM",
			MergeOpts{
				BaseBOM: &cdx.BOM{
					SpecVersion: cdx.SpecVersion1_6,
					Components: &[]cdx.Component{
						{Name: "comp1"},
					},
				},
			},
			"a452fb07f06a6aeda6167a7a117bf41073df4874287acaa4e0aaa1838ac1f80f"),
	)
})

var _ = Describe("StableBOMChecksum", func() {
	It("should return same checksum for BOMs with different SerialNumber but same content", func() {
		bom1 := &cdx.BOM{
			SerialNumber: "urn:uuid:11111111-1111-1111-1111-111111111111",
			Version:      1,
			Components: &[]cdx.Component{
				{Name: "test", Version: "1.0.0", Type: cdx.ComponentTypeLibrary},
			},
		}
		bom2 := &cdx.BOM{
			SerialNumber: "urn:uuid:22222222-2222-2222-2222-222222222222",
			Version:      2,
			Components: &[]cdx.Component{
				{Name: "test", Version: "1.0.0", Type: cdx.ComponentTypeLibrary},
			},
		}

		Expect(StableBOMChecksum(bom1)).To(Equal(StableBOMChecksum(bom2)))
	})

	It("should return different checksum for BOMs with different components", func() {
		bom1 := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "test1", Version: "1.0.0"},
			},
		}
		bom2 := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "test2", Version: "1.0.0"},
			},
		}

		Expect(StableBOMChecksum(bom1)).NotTo(Equal(StableBOMChecksum(bom2)))
	})

	It("should return empty string for nil BOM", func() {
		Expect(StableBOMChecksum(nil)).To(Equal(""))
	})

	It("should include services in checksum", func() {
		bom1 := &cdx.BOM{
			Services: &[]cdx.Service{
				{Name: "service1"},
			},
		}
		bom2 := &cdx.BOM{
			Services: &[]cdx.Service{
				{Name: "service2"},
			},
		}

		Expect(StableBOMChecksum(bom1)).NotTo(Equal(StableBOMChecksum(bom2)))
	})

	It("should include properties in checksum", func() {
		bom1 := &cdx.BOM{
			Properties: &[]cdx.Property{
				{Name: "prop1", Value: "value1"},
			},
		}
		bom2 := &cdx.BOM{
			Properties: &[]cdx.Property{
				{Name: "prop1", Value: "value2"},
			},
		}

		Expect(StableBOMChecksum(bom1)).NotTo(Equal(StableBOMChecksum(bom2)))
	})

	It("should include metadata in checksum", func() {
		bom1 := &cdx.BOM{
			Metadata: &cdx.Metadata{
				Component: &cdx.Component{Name: "metadata-comp-1"},
			},
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0.0"},
			},
		}
		bom2 := &cdx.BOM{
			Metadata: &cdx.Metadata{
				Component: &cdx.Component{Name: "metadata-comp-2"},
			},
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0.0"},
			},
		}

		Expect(StableBOMChecksum(bom1)).NotTo(Equal(StableBOMChecksum(bom2)))
	})

	It("should ignore signature differences", func() {
		bom1 := &cdx.BOM{
			Signature: &cdx.JSFSignature{
				JSFSigner: &cdx.JSFSigner{Algorithm: "RS256", Value: "sig-value"},
			},
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0.0"},
			},
		}
		bom2 := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0.0"},
			},
		}

		Expect(StableBOMChecksum(bom1)).To(Equal(StableBOMChecksum(bom2)))
	})

	It("should ignore metadata timestamp differences", func() {
		bom1 := &cdx.BOM{
			Metadata: &cdx.Metadata{
				Timestamp: "2024-01-01T00:00:00Z",
				Component: &cdx.Component{Name: "comp"},
			},
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0.0"},
			},
		}
		bom2 := &cdx.BOM{
			Metadata: &cdx.Metadata{
				Timestamp: "2025-06-15T12:30:00Z",
				Component: &cdx.Component{Name: "comp"},
			},
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0.0"},
			},
		}
		bom3 := &cdx.BOM{
			Metadata: &cdx.Metadata{
				Component: &cdx.Component{Name: "comp"},
			},
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0.0"},
			},
		}

		Expect(StableBOMChecksum(bom1)).To(Equal(StableBOMChecksum(bom2)))
		Expect(StableBOMChecksum(bom1)).To(Equal(StableBOMChecksum(bom3)))
	})

	It("should include vulnerabilities in checksum", func() {
		bom1 := &cdx.BOM{
			Vulnerabilities: &[]cdx.Vulnerability{
				{ID: "CVE-2024-0001"},
			},
		}
		bom2 := &cdx.BOM{
			Vulnerabilities: &[]cdx.Vulnerability{
				{ID: "CVE-2024-0002"},
			},
		}

		Expect(StableBOMChecksum(bom1)).NotTo(Equal(StableBOMChecksum(bom2)))
	})

	It("should include compositions in checksum", func() {
		bom1 := &cdx.BOM{
			Compositions: &[]cdx.Composition{
				{Aggregate: cdx.CompositionAggregateComplete},
			},
		}
		bom2 := &cdx.BOM{
			Compositions: &[]cdx.Composition{
				{Aggregate: cdx.CompositionAggregateIncomplete},
			},
		}

		Expect(StableBOMChecksum(bom1)).NotTo(Equal(StableBOMChecksum(bom2)))
	})
})
