package stage

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.DescribeTable("Git stage dependency defaults",
	func(declared map[StageName][]string, expected [][]string) {
		mapping := NewGitMapping()
		mapping.StagesDependencies = declared
		for i, name := range []StageName{Install, BeforeSetup, Setup} {
			gomega.Expect(mapping.stageDependenciesPaths(name)).To(gomega.Equal(expected[i]), string(name))
		}
	},
	ginkgo.Entry("absent block", nil, [][]string{{"**/*"}, {"**/*"}, {"**/*"}}),
	ginkgo.Entry("empty block", map[StageName][]string{}, [][]string{{"**/*"}, {"**/*"}, {"**/*"}}),
	ginkgo.Entry("beforeSetup masks", map[StageName][]string{BeforeSetup: {"src/**/*"}}, [][]string{nil, {"src/**/*"}, {"**/*"}}),
	ginkgo.Entry("install and setup masks", map[StageName][]string{Install: {"package.json", "package-lock.json"}, Setup: {"src/**/*"}}, [][]string{{"package.json", "package-lock.json"}, nil, {"src/**/*"}}),
	ginkgo.Entry("install masks", map[StageName][]string{Install: {"package-lock.json"}}, [][]string{{"package-lock.json"}, {"**/*"}, {"**/*"}}),
	ginkgo.Entry("explicit empty beforeSetup", map[StageName][]string{BeforeSetup: {}}, [][]string{nil, {}, {"**/*"}}),
	ginkgo.Entry("explicit empty setup", map[StageName][]string{Setup: {}}, [][]string{nil, nil, {}}),
	ginkgo.Entry("explicit empty install", map[StageName][]string{Install: {}}, [][]string{{}, {"**/*"}, {"**/*"}}),
	ginkgo.Entry("all stages declared", map[StageName][]string{Install: {"package.json"}, BeforeSetup: {"config/**/*"}, Setup: {"src/**/*"}}, [][]string{{"package.json"}, {"config/**/*"}, {"src/**/*"}}),
	ginkgo.Entry("last declaration wins over earlier one", map[StageName][]string{Install: {}, Setup: {}}, [][]string{{}, nil, {}}),
)
