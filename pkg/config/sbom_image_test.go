package config

import (
	"errors"

	cdx "github.com/CycloneDX/cyclonedx-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pkgsbom "github.com/werf/werf/v2/pkg/sbom"
)

func strPtr(s string) *string { return &s }

var _ = Describe("buildImageSbom", func() {
	DescribeTable("validate and build image sbom",
		func(meta *Meta, raw *rawSbom, d *doc, errMatcher OmegaMatcher, expectConfigErr bool, validate func(*Sbom)) {
			sbomDirective, err := buildImageSbom(meta, raw, d)
			Expect(err).To(errMatcher)

			if err != nil {
				if expectConfigErr {
					var confErr *configError
					Expect(errors.As(err, &confErr)).To(BeTrue())
				}
				return
			}

			if validate != nil {
				validate(sbomDirective)
			}
		},
		Entry(
			"should fail when build.sbom.enable=false and image sbom is specified",
			&Meta{
				Build: MetaBuild{
					Sbom: &MetaBuildSbom{
						Enable:   false,
						Standard: pkgsbom.StandardTypeCycloneDX16,
					},
				},
			},
			&rawSbom{
				Fragment: strPtr("components: []"),
			},
			&doc{RenderFilePath: "werf.yaml", Content: []byte("image: test")},
			HaveOccurred(),
			true,
			nil,
		),
		Entry(
			"should succeed when build.sbom.enable=true and image sbom is not specified (now optional)",
			&Meta{
				Build: MetaBuild{
					Sbom: &MetaBuildSbom{
						Enable:   true,
						Standard: pkgsbom.StandardTypeCycloneDX16,
					},
				},
			},
			nil,
			&doc{RenderFilePath: "werf.yaml", Content: []byte("image: test")},
			Succeed(),
			false,
			func(sbomDirective *Sbom) {
				Expect(sbomDirective).To(BeNil())
			},
		),
		Entry(
			"should fail when build.sbom.enable=true and sbom.fragment is empty",
			&Meta{
				Build: MetaBuild{
					Sbom: &MetaBuildSbom{
						Enable:   true,
						Standard: pkgsbom.StandardTypeCycloneDX16,
					},
				},
			},
			&rawSbom{
				Fragment: strPtr("   "),
			},
			&doc{RenderFilePath: "werf.yaml", Content: []byte("image: test")},
			HaveOccurred(),
			true,
			nil,
		),
		Entry(
			"should succeed when build.sbom.enable=true and sbom.fragment contains valid YAML (components)",
			&Meta{
				Build: MetaBuild{
					Sbom: &MetaBuildSbom{
						Enable:   true,
						Standard: pkgsbom.StandardTypeCycloneDX16,
					},
				},
			},
			&rawSbom{
				Fragment: strPtr(`
components:
  - type: library
    name: openssl
    version: "3.0.0"
`),
			},
			&doc{RenderFilePath: "werf.yaml", Content: []byte("image: test")},
			Succeed(),
			false,
			func(sbomDirective *Sbom) {
				Expect(sbomDirective).ToNot(BeNil())
				Expect(sbomDirective.Document).ToNot(BeNil())

				Expect(sbomDirective.Document.BOMFormat).To(Equal(cdx.BOMFormat))
				Expect(sbomDirective.Document.SpecVersion).To(Equal(cdx.SpecVersion1_6))
				Expect(sbomDirective.Document.Version).To(BeNumerically(">=", 1))

				Expect(sbomDirective.Document.Components).ToNot(BeNil())
				Expect(*sbomDirective.Document.Components).To(HaveLen(1))
				Expect((*sbomDirective.Document.Components)[0].Name).To(Equal("openssl"))
			},
		),
	)
})
