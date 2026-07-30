package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/werf/common-go/pkg/util"
)

var _ = Describe("prepareWerfConfig image names", func() {
	var giterminismManager *GiterminismManagerStub

	BeforeEach(func() {
		parentStack = util.NewStack()
		giterminismManager = NewGiterminismManagerStub(NewLocalGitRepoStub("9d8059842b6fde712c58315ca0ab4713d90761c0"))
	})

	prepare := func(imageName string) error {
		rawYaml, err := yaml.Marshal(map[string]interface{}{
			"image": imageName,
			"from":  "ubuntu:22.04",
		})
		Expect(err).To(Succeed())

		doc := &doc{Content: rawYaml}
		rawImage := &rawStapelImage{doc: doc}
		Expect(yaml.UnmarshalStrict(doc.Content, rawImage)).To(Succeed())

		meta := &Meta{}
		meta.ConfigVersion = 1
		meta.Project = "test"

		_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage}, nil, meta)
		return err
	}

	DescribeTable("rejects a name that cannot have been meant", func(imageName string) {
		Expect(prepare(imageName)).To(MatchError(ContainSubstring("should comply with regex")))
	},
		Entry("empty", ""),
		Entry("empty segment", "app//api"),
		Entry("leading separator", "/api"),
		Entry("trailing separator", "api-"),
		Entry("whitespace", "my app"),
	)

	DescribeTable("accepts a valid name", func(imageName string) {
		Expect(prepare(imageName)).To(Succeed())
	},
		Entry("plain", "app"),
		Entry("hierarchical", "modules/api"),
		Entry("underscores and case", "Dockerfile_base_image"),
	)
})
