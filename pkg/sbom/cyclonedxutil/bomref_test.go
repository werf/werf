package cyclonedxutil

import (
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func collectBOMRefs(bom *cdx.BOM) []string {
	var refs []string

	if bom.Components != nil {
		for _, c := range *bom.Components {
			if c.BOMRef != "" {
				refs = append(refs, c.BOMRef)
			}
		}
	}

	if bom.Services != nil {
		for _, s := range *bom.Services {
			if s.BOMRef != "" {
				refs = append(refs, s.BOMRef)
			}
		}
	}

	return refs
}

func uniqueStrings(ss []string) bool {
	seen := map[string]struct{}{}
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			return false
		}
		seen[s] = struct{}{}
	}

	return true
}

var _ = Describe("deriveBomRef", func() {
	DescribeTable("generates correct bom-ref",
		func(purl, serial string, index int, check func(string)) {
			check(deriveBomRef(purl, serial, index))
		},

		Entry("appends package-id qualifier to valid PURL",
			"pkg:deb/debian/curl@7.74.0", "urn:uuid:test-serial", 0,
			func(ref string) {
				Expect(ref).To(HavePrefix("pkg:deb/debian/curl@7.74.0"))
				Expect(ref).To(ContainSubstring("package-id="))
			},
		),
		Entry("preserves existing PURL qualifiers",
			"pkg:deb/debian/curl@7.74.0?arch=amd64", "urn:uuid:test-serial", 0,
			func(ref string) {
				Expect(ref).To(ContainSubstring("arch=amd64"))
				Expect(ref).To(ContainSubstring("package-id="))
			},
		),
		Entry("returns raw ID for empty PURL",
			"", "urn:uuid:test-serial", 0,
			func(ref string) {
				Expect(ref).ToNot(BeEmpty())
				Expect(strings.HasPrefix(ref, "pkg:")).To(BeFalse())
			},
		),
		Entry("returns raw ID for invalid PURL",
			"not-a-purl", "urn:uuid:test-serial", 0,
			func(ref string) {
				Expect(ref).ToNot(BeEmpty())
				Expect(strings.HasPrefix(ref, "pkg:")).To(BeFalse())
			},
		),
	)

	It("generates different refs for different indices", func() {
		Expect(
			deriveBomRef("pkg:deb/curl@8.12", "urn:uuid:serial", 0),
		).ToNot(Equal(
			deriveBomRef("pkg:deb/curl@8.12", "urn:uuid:serial", 1),
		))
	})

	It("generates different refs for different serials", func() {
		Expect(
			deriveBomRef("pkg:deb/curl@8.12", "urn:uuid:serial-a", 0),
		).ToNot(Equal(
			deriveBomRef("pkg:deb/curl@8.12", "urn:uuid:serial-b", 0),
		))
	})
})

var _ = Describe("ensureUniqueBOMRefs", func() {
	It("handles nil BOM without panic", func() {
		Expect(func() { ensureUniqueBOMRefs(nil) }).ToNot(Panic())
	})

	It("is idempotent", func() {
		bom := &cdx.BOM{
			SerialNumber: "urn:uuid:test",
			Components: &[]cdx.Component{
				{BOMRef: "curl", PackageURL: "pkg:deb/curl@8.12", Name: "curl", Version: "8.12"},
			},
		}
		ensureUniqueBOMRefs(bom)
		first := collectBOMRefs(bom)
		ensureUniqueBOMRefs(bom)
		Expect(collectBOMRefs(bom)).To(Equal(first))
	})

	It("skips components with empty bom-ref", func() {
		bom := &cdx.BOM{
			SerialNumber: "urn:uuid:test",
			Components: &[]cdx.Component{
				{Name: "no-ref-comp", Version: "1.0"},
			},
		}
		ensureUniqueBOMRefs(bom)
		Expect((*bom.Components)[0].BOMRef).To(BeEmpty())
	})

	DescribeTable("assigns unique bom-refs",
		func(bom *cdx.BOM, expectedRefCount int, expectPackageID bool) {
			ensureUniqueBOMRefs(bom)
			refs := collectBOMRefs(bom)
			Expect(refs).To(HaveLen(expectedRefCount))
			Expect(uniqueStrings(refs)).To(BeTrue())
			if expectPackageID {
				for _, r := range refs {
					Expect(r).To(ContainSubstring("package-id="))
				}
			}
		},

		Entry("components with PURL get package-id in bom-ref",
			&cdx.BOM{
				SerialNumber: "urn:uuid:test",
				Components: &[]cdx.Component{
					{BOMRef: "curl", PackageURL: "pkg:deb/curl@7.74", Name: "curl", Version: "7.74"},
					{BOMRef: "curl", PackageURL: "pkg:deb/curl@8.12", Name: "curl", Version: "8.12"},
				},
			}, 2, true,
		),
		Entry("components without PURL get unique fallback refs",
			&cdx.BOM{
				SerialNumber: "urn:uuid:test",
				Components: &[]cdx.Component{
					{BOMRef: "comp", Name: "alpha", Version: "1.0"},
					{BOMRef: "comp", Name: "beta", Version: "1.0"},
				},
			}, 2, false,
		),
		Entry("components with same PURL are kept (no identity dedup)",
			&cdx.BOM{
				SerialNumber: "urn:uuid:test",
				Components: &[]cdx.Component{
					{BOMRef: "A-curl", PackageURL: "pkg:deb/curl@8.12", Name: "curl", Version: "8.12"},
					{BOMRef: "B-curl", PackageURL: "pkg:deb/curl@8.12", Name: "curl", Version: "8.12"},
				},
			}, 2, true,
		),
		Entry("services get unique refs",
			&cdx.BOM{
				SerialNumber: "urn:uuid:test",
				Services: &[]cdx.Service{
					{BOMRef: "svc", Name: "svc-a", Version: "1.0"},
					{BOMRef: "svc", Name: "svc-b", Version: "2.0"},
				},
			}, 2, false,
		),
	)

	DescribeTable("rewrites cross-references after bom-ref regeneration",
		func(bom *cdx.BOM, check func(*cdx.BOM)) {
			ensureUniqueBOMRefs(bom)
			check(bom)
		},

		Entry("dependency refs point to new component bom-refs",
			&cdx.BOM{
				SerialNumber: "urn:uuid:test",
				Components: &[]cdx.Component{
					{BOMRef: "old-curl", PackageURL: "pkg:deb/curl@8.12", Name: "curl", Version: "8.12"},
					{BOMRef: "old-zlib", PackageURL: "pkg:deb/zlib@1.2", Name: "zlib", Version: "1.2"},
				},
				Dependencies: &[]cdx.Dependency{
					{Ref: "old-curl", Dependencies: &[]string{"old-zlib"}},
				},
			},
			func(bom *cdx.BOM) {
				newCurlRef := (*bom.Components)[0].BOMRef
				newZlibRef := (*bom.Components)[1].BOMRef
				dep := (*bom.Dependencies)[0]
				Expect(dep.Ref).To(Equal(newCurlRef))
				Expect((*dep.Dependencies)[0]).To(Equal(newZlibRef))
			},
		),
		Entry("vulnerability affects refs point to new component bom-refs",
			&cdx.BOM{
				SerialNumber: "urn:uuid:test",
				Components: &[]cdx.Component{
					{BOMRef: "old-curl", PackageURL: "pkg:deb/curl@8.12", Name: "curl", Version: "8.12"},
				},
				Vulnerabilities: &[]cdx.Vulnerability{
					{BOMRef: "CVE-1", ID: "CVE-2024-0001", Affects: &[]cdx.Affects{{Ref: "old-curl"}}},
				},
			},
			func(bom *cdx.BOM) {
				newCurlRef := (*bom.Components)[0].BOMRef
				Expect(newCurlRef).ToNot(Equal("old-curl"))
				Expect((*(*bom.Vulnerabilities)[0].Affects)[0].Ref).To(Equal(newCurlRef))
			},
		),
	)
})
