package cyclonedxutil

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
			func(fragmentYAML string, expectedErrMatcher types.GomegaMatcher, expectedBom bomAssert) {
				bom, err := BuildCycloneDX16BOMFromYAMLFragment([]byte(fragmentYAML))

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
				"invalid YAML",
				"a: [",
				MatchError(ContainSubstring("invalid YAML fragment")),
				nil,
			),

			Entry(
				"succeeds for a valid fragment (components)",
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
})
