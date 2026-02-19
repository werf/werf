package sbom

import (
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

type bomAssert func(*cdx.BOM)

var _ = Describe("SBOM CycloneDX builders", func() {
	Describe("BuildCycloneDX16BOMFromYAMLFragment", func() {
		DescribeTable(
			"builds/validates BOM from YAML fragment",
			func(standard StandardType, fragmentYAML string, expectedErrMatcher types.GomegaMatcher, expectedBom bomAssert) {
				bom, err := BuildCycloneDX16BOMFromYAMLFragment(standard, []byte(fragmentYAML))

				Expect(err).To(expectedErrMatcher)
				if err != nil {
					return
				}

				Expect(bom).ToNot(BeNil())
				Expect(bom.BOMFormat).To(Equal(cdx.BOMFormat))
				Expect(bom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
				Expect(bom.Version).To(BeNumerically(">=", 1))

				if expectedBom != nil {
					expectedBom(bom)
				}
			},

			Entry(
				"unsupported standard",
				StandardTypeSPDX23,
				"components: []",
				MatchError(ContainSubstring("unsupported standard")),
				nil,
			),
			Entry(
				"empty fragment",
				StandardTypeCycloneDX16,
				"   \n\t",
				MatchError(ContainSubstring("document fragment is empty")),
				nil,
			),
			Entry(
				"invalid YAML",
				StandardTypeCycloneDX16,
				"a: [",
				MatchError(ContainSubstring("invalid YAML fragment")),
				nil,
			),

			Entry(
				"overrides specVersion to unsupported value",
				StandardTypeCycloneDX16,
				"specVersion: \"1.5\"\n",
				MatchError(ContainSubstring("invalid specVersion")),
				nil,
			),
			Entry(
				"overrides bomFormat to unsupported value",
				StandardTypeCycloneDX16,
				"bomFormat: \"SPDX\"\n",
				MatchError(ContainSubstring("invalid bomFormat")),
				nil,
			),
			Entry(
				"sets invalid serialNumber",
				StandardTypeCycloneDX16,
				"serialNumber: \"not-a-urn\"\n",
				MatchError(ContainSubstring("serialNumber")),
				nil,
			),
			Entry(
				"succeeds for a valid fragment (components)",
				StandardTypeCycloneDX16,
				`
components:
  - type: library
    name: openssl
    version: "3.0.0"
`,
				Succeed(),
				bomAssert(func(b *cdx.BOM) {
					Expect(b.SerialNumber).To(HavePrefix("urn:uuid:"))
					_, parseErr := uuid.Parse(strings.TrimPrefix(b.SerialNumber, "urn:uuid:"))
					Expect(parseErr).ToNot(HaveOccurred())

					Expect(b.Components).ToNot(BeNil())
					Expect(*b.Components).To(HaveLen(1))
					Expect((*b.Components)[0].Name).To(Equal("openssl"))
					Expect((*b.Components)[0].Type).To(Equal(cdx.ComponentTypeLibrary))
					Expect((*b.Components)[0].Version).To(Equal("3.0.0"))
				}),
			),
			Entry(
				"allows overriding base fields via full document YAML (version and serialNumber)",
				StandardTypeCycloneDX16,
				`
bomFormat: CycloneDX
specVersion: "1.6"
version: 7
serialNumber: "urn:uuid:00000000-0000-0000-0000-000000000000"
components:
  - type: library
    name: zlib
`,
				Succeed(),
				bomAssert(func(b *cdx.BOM) {
					Expect(b.Version).To(Equal(7))
					Expect(b.SerialNumber).To(Equal("urn:uuid:00000000-0000-0000-0000-000000000000"))
				}),
			),
		)
	})

	Describe("BuildCycloneDX16BOMFromJSON", func() {
		DescribeTable(
			"builds/validates BOM from JSON bytes",
			func(standard StandardType, bomJSON string, expectedErrMatcher types.GomegaMatcher, expectedBom bomAssert) {
				bom, err := BuildCycloneDX16BOMFromJSON(standard, []byte(bomJSON))

				Expect(err).To(expectedErrMatcher)
				if err != nil {
					return
				}

				Expect(bom).ToNot(BeNil())
				Expect(bom.BOMFormat).To(Equal(cdx.BOMFormat))
				Expect(bom.SpecVersion).To(Equal(cdx.SpecVersion1_6))
				Expect(bom.Version).To(BeNumerically(">=", 1))

				if expectedBom != nil {
					expectedBom(bom)
				}
			},

			Entry(
				"unsupported standard",
				StandardTypeSPDX23,
				`{"bomFormat":"CycloneDX","specVersion":"1.6","version":1}`,
				MatchError(ContainSubstring("unsupported standard")),
				nil,
			),
			Entry(
				"empty json",
				StandardTypeCycloneDX16,
				"  \n\t",
				MatchError(ContainSubstring("document json is empty")),
				nil,
			),
			Entry(
				"invalid json syntax",
				StandardTypeCycloneDX16,
				`{"bomFormat":`,
				MatchError(ContainSubstring("failed to decode")),
				nil,
			),
			Entry(
				"invalid bomFormat",
				StandardTypeCycloneDX16,
				`{"bomFormat":"SPDX","specVersion":"1.6","version":1}`,
				MatchError(ContainSubstring("invalid bomFormat")),
				nil,
			),
			Entry(
				"invalid specVersion",
				StandardTypeCycloneDX16,
				`{"bomFormat":"CycloneDX","specVersion":"1.5","version":1}`,
				MatchError(ContainSubstring("invalid specVersion")),
				nil,
			),
			Entry(
				"missing specVersion (zero value is not 1.6)",
				StandardTypeCycloneDX16,
				`{"bomFormat":"CycloneDX","version":1}`,
				MatchError(ContainSubstring("invalid specVersion")),
				nil,
			),
			Entry(
				"version must be >= 1",
				StandardTypeCycloneDX16,
				`{"bomFormat":"CycloneDX","specVersion":"1.6","version":0}`,
				MatchError(ContainSubstring("version must be >= 1")),
				nil,
			),
			Entry(
				"invalid serialNumber format",
				StandardTypeCycloneDX16,
				`{"bomFormat":"CycloneDX","specVersion":"1.6","version":1,"serialNumber":"not-a-urn"}`,
				MatchError(ContainSubstring("serialNumber")),
				nil,
			),
			Entry(
				"succeeds for a valid BOM JSON",
				StandardTypeCycloneDX16,
				`{
  "bomFormat": "CycloneDX",
  "specVersion": "1.6",
  "version": 3,
  "serialNumber": "urn:uuid:00000000-0000-0000-0000-000000000000",
  "components": [
    { "type": "library", "name": "openssl", "version": "3.0.0" }
  ]
}`,
				Succeed(),
				bomAssert(func(b *cdx.BOM) {
					Expect(b.Version).To(Equal(3))
					Expect(b.SerialNumber).To(Equal("urn:uuid:00000000-0000-0000-0000-000000000000"))

					Expect(b.Components).ToNot(BeNil())
					Expect(*b.Components).To(HaveLen(1))
					Expect((*b.Components)[0].Name).To(Equal("openssl"))
					Expect((*b.Components)[0].Type).To(Equal(cdx.ComponentTypeLibrary))
					Expect((*b.Components)[0].Version).To(Equal("3.0.0"))
				}),
			),
			Entry(
				"succeeds when serialNumber is omitted (it is optional for validation)",
				StandardTypeCycloneDX16,
				`{
  "bomFormat": "CycloneDX",
  "specVersion": "1.6",
  "version": 1
}`,
				Succeed(),
				bomAssert(func(b *cdx.BOM) {
					Expect(b.Version).To(Equal(1))
					Expect(b.SerialNumber).To(BeEmpty())
				}),
			),
		)
	})
})
