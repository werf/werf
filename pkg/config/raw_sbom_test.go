package config

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/werf/common-go/pkg/util"
)

var _ = Describe("rawSbom (YAML-level validation)", func() {
	DescribeTable(
		"fragment validation when sbom section is present",
		func(yamlMap map[string]interface{}, expectedSbomPresent bool, unmarshalMatcher, configErrMatcher OmegaMatcher) {
			// NOTE: global var used by UnmarshalYAML parent tracking across many config raw structs.
			parentStack = util.NewStack()

			rawYaml, err := yaml.Marshal(yamlMap)
			Expect(err).To(Succeed())

			d := &doc{
				RenderFilePath: "werf.yaml",
				Content:        rawYaml,
			}

			rawDockerfileImage := &rawImageFromDockerfile{doc: d}
			err = yaml.UnmarshalStrict(d.Content, rawDockerfileImage)

			Expect(err).To(unmarshalMatcher)

			if err != nil {
				var confErr *configError
				Expect(errors.As(err, &confErr)).To(configErrMatcher)
				return
			}

			Expect(rawDockerfileImage).ToNot(BeNil())
			Expect(rawDockerfileImage.RawSbom != nil).To(Equal(expectedSbomPresent))
		},

		Entry(
			"should succeed when sbom section is omitted",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
			},
			false,
			Succeed(),
			BeFalse(),
		),

		Entry(
			"should fail when sbom section exists but fragment is missing",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"sbom":       map[string]interface{}{},
			},
			false,
			HaveOccurred(),
			BeTrue(),
		),

		Entry(
			"should fail when sbom.fragment is empty",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"sbom": map[string]interface{}{
					"fragment": "   ",
				},
			},
			false,
			HaveOccurred(),
			BeTrue(),
		),

		Entry(
			"should fail when sbom.fragment is not valid YAML",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"sbom": map[string]interface{}{
					"fragment": "components: [",
				},
			},
			false,
			HaveOccurred(),
			BeTrue(),
		),

		Entry(
			"should fail when sbom.fragment YAML root is not a mapping",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"sbom": map[string]interface{}{
					"fragment": "- a\n- b\n",
				},
			},
			false,
			HaveOccurred(),
			BeTrue(),
		),

		Entry(
			"should succeed when sbom.fragment contains valid YAML mapping",
			map[string]interface{}{
				"image":      "image1",
				"dockerfile": "Dockerfile",
				"sbom": map[string]interface{}{
					"fragment": "components: []\n",
				},
			},
			true,
			Succeed(),
			BeFalse(),
		),
	)
})
