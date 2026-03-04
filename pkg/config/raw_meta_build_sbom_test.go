package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/sbom"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

var _ = Describe("rawMetaBuildSbom", func() {
	BeforeEach(func() {
		// NOTE: global var used by UnmarshalYAML parent tracking across many config raw structs.
		parentStack = util.NewStack()
	})

	DescribeTable("unmarshal and convert to directive",
		func(yamlMap map[string]interface{}, expected *MetaBuildSbom, unmarshalMatcher OmegaMatcher) {
			// We intentionally unmarshal only the sbom section itself (not the whole meta build yaml),
			// because `rawMetaBuildSbom` is the structure representing sbom subsection.
			//
			// NOTE: empty map is a valid case (means "sbom section exists but has no fields"),
			// and should trigger defaults.
			rawYaml, err := yaml.Marshal(yamlMap)
			Expect(err).To(Succeed())

			doc := &doc{Content: rawYaml}

			var rawSbom rawMetaBuildSbom
			rawSbom.rawMetaBuild = &rawMetaBuild{
				rawMeta: &rawMeta{doc: doc},
			}

			err = yaml.UnmarshalStrict(doc.Content, &rawSbom)
			Expect(err).To(unmarshalMatcher)

			if err != nil {
				return
			}

			Expect(rawSbom.toDirective()).To(Equal(expected))
		},
		Entry(
			"should apply defaults when enable and standard are omitted",
			map[string]interface{}{},
			&MetaBuildSbom{
				Enable:   false,
				Standard: sbom.StandardTypeCycloneDX16,
				Gost:     gost.DefaultConfig(),
			},
			Succeed(),
		),
		Entry(
			"should require standard when enable=true",
			map[string]interface{}{
				"enable": true,
			},
			nil,
			HaveOccurred(),
		),
		Entry(
			"should not require standard when enable=false (standard defaults to cyclonedx@1.6)",
			map[string]interface{}{
				"enable": false,
			},
			&MetaBuildSbom{
				Enable:   false,
				Standard: sbom.StandardTypeCycloneDX16,
				Gost:     gost.DefaultConfig(),
			},
			Succeed(),
		),
		Entry(
			"should accept enable=true with standard=cyclonedx@1.6",
			map[string]interface{}{
				"enable":   true,
				"standard": "cyclonedx@1.6",
			},
			&MetaBuildSbom{
				Enable:   true,
				Standard: sbom.StandardTypeCycloneDX16,
				Gost:     gost.DefaultConfig(),
			},
			Succeed(),
		),
		Entry(
			"should accept enable=false with standard=cyclonedx@1.6",
			map[string]interface{}{
				"enable":   false,
				"standard": "cyclonedx@1.6",
			},
			&MetaBuildSbom{
				Enable:   false,
				Standard: sbom.StandardTypeCycloneDX16,
				Gost:     gost.DefaultConfig(),
			},
			Succeed(),
		),
		Entry(
			"should reject standard=cyclonedx@1.6 without enable (enable must be explicitly set to true)",
			map[string]interface{}{
				"standard": "cyclonedx@1.6",
			},
			nil,
			HaveOccurred(),
		),
		Entry(
			"should reject unsupported standard values",
			map[string]interface{}{
				"standard": "cyclonedx@1.5",
			},
			nil,
			HaveOccurred(),
		),
		Entry(
			"should accept gost section even when enable=false",
			map[string]interface{}{
				"enable": false,
				"gost": map[string]interface{}{
					"attackSurface": "yes",
				},
			},
			&MetaBuildSbom{
				Enable:   false,
				Standard: sbom.StandardTypeCycloneDX16,
				Gost: gost.Config{
					AttackSurface:    gost.GostValueYes,
					SecurityFunction: gost.GostValueYes,
				},
			},
			Succeed(),
		),
		Entry(
			"should accept gost with inherit",
			map[string]interface{}{
				"enable":   true,
				"standard": "cyclonedx@1.6",
				"gost": map[string]interface{}{
					"attackSurface": "inherit",
				},
			},
			&MetaBuildSbom{
				Enable:   true,
				Standard: sbom.StandardTypeCycloneDX16,
				Gost: gost.Config{
					AttackSurface:    gost.GostValueInherit,
					SecurityFunction: gost.GostValueYes,
				},
			},
			Succeed(),
		),
		Entry(
			"should reject invalid gost values",
			map[string]interface{}{
				"gost": map[string]interface{}{
					"attackSurface": "invalid",
				},
			},
			nil,
			HaveOccurred(),
		),
	)
})
