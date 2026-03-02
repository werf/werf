package cyclonedxutil

import (
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		Expect(result.Dependencies).To(BeNil())
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

	It("does not deduplicate components", func() {
		duplicateComp := cdx.Component{Name: "duplicate-comp", Version: "1.0.0"}
		baseBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{duplicateComp}}
		targetBOM := &cdx.BOM{SpecVersion: cdx.SpecVersion1_6, Components: &[]cdx.Component{duplicateComp}}

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
})

var _ = Describe("BOMChecksum", func() {
	It("returns consistent hash for same BOM", func() {
		bom := &cdx.BOM{
			BOMFormat:   cdx.BOMFormat,
			SpecVersion: cdx.SpecVersion1_6,
			Version:     1,
			Components: &[]cdx.Component{
				{Name: "test-comp", Version: "1.0.0"},
			},
		}

		checksum1 := BOMChecksum(bom)
		checksum2 := BOMChecksum(bom)

		Expect(checksum1).ToNot(BeEmpty())
		Expect(checksum1).To(Equal(checksum2))
	})

	It("is different for different BOMs", func() {
		bom1 := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "comp-1", Version: "1.0.0"},
			},
		}
		bom2 := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "comp-2", Version: "1.0.0"},
			},
		}

		Expect(BOMChecksum(bom1)).ToNot(Equal(BOMChecksum(bom2)))
	})

	It("returns empty for nil", func() {
		Expect(BOMChecksum(nil)).To(BeEmpty())
	})
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

	It("Checksum: same opts -> same checksum; different opts -> different checksum", func() {
		bom1 := &cdx.BOM{Components: &[]cdx.Component{{Name: "comp1"}}}
		bom2 := &cdx.BOM{Components: &[]cdx.Component{{Name: "comp2"}}}

		opts1 := MergeOpts{BaseBOM: bom1}
		opts2 := MergeOpts{BaseBOM: bom2}
		opts3 := MergeOpts{BaseBOM: bom1}

		checksum1 := opts1.Checksum()
		checksum2 := opts2.Checksum()
		checksum3 := opts3.Checksum()

		Expect(checksum1).ToNot(BeEmpty())
		Expect(checksum1).ToNot(Equal(checksum2))
		Expect(checksum1).To(Equal(checksum3))
	})

	It("Checksum is empty for empty opts", func() {
		Expect(MergeOpts{}.Checksum()).To(BeEmpty())
	})
})

var _ = Describe("internal sanity", func() {
	It("strings import is used", func() {
		Expect(strings.HasPrefix("urn:uuid:x", "urn:uuid:")).To(BeTrue())
	})
})
