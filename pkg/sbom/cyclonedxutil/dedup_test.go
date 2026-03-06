package cyclonedxutil

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dedupJSONSlice", func() {
	It("removes identical components", func() {
		items := []cdx.Component{
			{Name: "a", Version: "1.0"},
			{Name: "b", Version: "2.0"},
			{Name: "a", Version: "1.0"},
		}
		result := dedupJSONSlice(items)
		Expect(result).To(HaveLen(2))
		Expect(result[0].Name).To(Equal("a"))
		Expect(result[1].Name).To(Equal("b"))
	})

	It("keeps components that differ by any field", func() {
		items := []cdx.Component{
			{Name: "a", Version: "1.0"},
			{Name: "a", Version: "2.0"},
		}
		result := dedupJSONSlice(items)
		Expect(result).To(HaveLen(2))
	})

	It("returns empty slice unchanged", func() {
		result := dedupJSONSlice([]cdx.Component{})
		Expect(result).To(BeEmpty())
	})

	It("returns nil slice unchanged", func() {
		result := dedupJSONSlice[cdx.Component](nil)
		Expect(result).To(BeNil())
	})

	It("preserves order (first occurrence wins)", func() {
		items := []cdx.Component{
			{Name: "c", Version: "1.0"},
			{Name: "a", Version: "1.0"},
			{Name: "b", Version: "1.0"},
			{Name: "a", Version: "1.0"},
			{Name: "c", Version: "1.0"},
		}
		result := dedupJSONSlice(items)
		Expect(result).To(HaveLen(3))
		Expect(result[0].Name).To(Equal("c"))
		Expect(result[1].Name).To(Equal("a"))
		Expect(result[2].Name).To(Equal("b"))
	})

	It("deduplicates dependencies", func() {
		items := []cdx.Dependency{
			{Ref: "pkg:npm/lodash@4.17.21", Dependencies: &[]string{"dep-a"}},
			{Ref: "pkg:npm/lodash@4.17.21", Dependencies: &[]string{"dep-a"}},
		}
		result := dedupJSONSlice(items)
		Expect(result).To(HaveLen(1))
	})

	It("deduplicates vulnerabilities", func() {
		items := []cdx.Vulnerability{
			{ID: "CVE-2024-0001", Description: "test"},
			{ID: "CVE-2024-0001", Description: "test"},
			{ID: "CVE-2024-0002", Description: "other"},
		}
		result := dedupJSONSlice(items)
		Expect(result).To(HaveLen(2))
	})
})

var _ = Describe("dedupPtrSlice", func() {
	It("returns nil for nil input", func() {
		result := dedupPtrSlice[cdx.Component](nil)
		Expect(result).To(BeNil())
	})

	It("deduplicates and returns pointer", func() {
		items := &[]cdx.Component{
			{Name: "a", Version: "1.0"},
			{Name: "a", Version: "1.0"},
			{Name: "b", Version: "2.0"},
		}
		result := dedupPtrSlice(items)
		Expect(result).ToNot(BeNil())
		Expect(*result).To(HaveLen(2))
	})
})

var _ = Describe("DedupBOM", func() {
	It("deduplicates all uniqueItems sections", func() {
		bom := &cdx.BOM{
			Components: &[]cdx.Component{
				{Name: "comp", Version: "1.0"},
				{Name: "comp", Version: "1.0"},
			},
			Services: &[]cdx.Service{
				{Name: "svc"},
				{Name: "svc"},
			},
			Dependencies: &[]cdx.Dependency{
				{Ref: "ref-1"},
				{Ref: "ref-1"},
			},
			Vulnerabilities: &[]cdx.Vulnerability{
				{ID: "CVE-1"},
				{ID: "CVE-1"},
			},
			Compositions: &[]cdx.Composition{
				{Aggregate: cdx.CompositionAggregateComplete},
				{Aggregate: cdx.CompositionAggregateComplete},
			},
			Annotations: &[]cdx.Annotation{
				{Text: "note"},
				{Text: "note"},
			},
			Formulation: &[]cdx.Formula{
				{BOMRef: "f1"},
				{BOMRef: "f1"},
			},
		}

		DedupBOM(bom)

		Expect(*bom.Components).To(HaveLen(1))
		Expect(*bom.Services).To(HaveLen(1))
		Expect(*bom.Dependencies).To(HaveLen(1))
		Expect(*bom.Vulnerabilities).To(HaveLen(1))
		Expect(*bom.Compositions).To(HaveLen(1))
		Expect(*bom.Annotations).To(HaveLen(1))
		Expect(*bom.Formulation).To(HaveLen(1))
	})

	It("handles nil BOM", func() {
		Expect(func() { DedupBOM(nil) }).ToNot(Panic())
	})

	It("handles BOM with nil sections", func() {
		bom := &cdx.BOM{}
		DedupBOM(bom)
		Expect(bom.Components).To(BeNil())
		Expect(bom.Services).To(BeNil())
	})
})
