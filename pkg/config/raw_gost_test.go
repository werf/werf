package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/sbom/cyclonedxutil/gost"
)

var _ = Describe("rawGost", func() {
	DescribeTable("unmarshal and validate",
		func(yamlMap map[string]interface{}, expected gost.Config, unmarshalMatcher OmegaMatcher) {
			// NOTE: global var used by UnmarshalYAML parent tracking across many config raw structs.
			parentStack = util.NewStack()

			rawYaml, err := yaml.Marshal(yamlMap)
			Expect(err).To(Succeed())

			doc := &doc{Content: rawYaml}

			var raw rawGost
			raw.doc = doc

			err = yaml.UnmarshalStrict(doc.Content, &raw)
			Expect(err).To(unmarshalMatcher)

			if err != nil {
				return
			}

			Expect(raw.toConfig()).To(Equal(expected))
		},
		Entry("valid yes/no",
			map[string]interface{}{
				"attackSurface":    "yes",
				"securityFunction": "no",
			},
			gost.Config{AttackSurface: gost.GostValueYes, SecurityFunction: gost.GostValueNo},
			Succeed()),
		Entry("valid indirect",
			map[string]interface{}{
				"attackSurface": "indirect",
			},
			gost.Config{AttackSurface: gost.GostValueIndirect},
			Succeed()),
		Entry("invalid value",
			map[string]interface{}{
				"attackSurface": "maybe",
			},
			gost.Config{},
			HaveOccurred()),
	)
})
