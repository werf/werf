package config

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("rawStageDependencies", func() {
	ginkgo.DescribeTable("preserves the difference between an absent stage and an explicit empty list",
		func(gitYaml string, install, beforeSetup, setup []string) {
			gitLocals, err := parseStapelImageGitLocals(stapelImageWithGit(gitYaml))
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(gitLocals).To(gomega.HaveLen(1))

			stageDependencies := gitLocals[0].StageDependencies
			gomega.Expect(stageDependencies).NotTo(gomega.BeNil())

			gomega.Expect(stageDependencies.Install).To(gomega.Equal(install))
			gomega.Expect(stageDependencies.BeforeSetup).To(gomega.Equal(beforeSetup))
			gomega.Expect(stageDependencies.Setup).To(gomega.Equal(setup))
		},
		ginkgo.Entry("only beforeSetup masks",
			"- add: /\n  to: /app\n  stageDependencies:\n    beforeSetup: [\"src/**/*\"]\n",
			nil, []string{"src/**/*"}, nil,
		),
		ginkgo.Entry("install and setup masks",
			"- add: /\n  to: /app\n  stageDependencies:\n    install: [\"package.json\", \"package-lock.json\"]\n    setup: [\"src/**/*\"]\n",
			[]string{"package.json", "package-lock.json"}, nil, []string{"src/**/*"},
		),
		ginkgo.Entry("only install masks",
			"- add: /\n  to: /app\n  stageDependencies:\n    install: [\"package-lock.json\"]\n",
			[]string{"package-lock.json"}, nil, nil,
		),
		ginkgo.Entry("explicit empty beforeSetup",
			"- add: /\n  to: /app\n  stageDependencies:\n    beforeSetup: []\n",
			nil, []string{}, nil,
		),
		ginkgo.Entry("explicit empty setup",
			"- add: /\n  to: /app\n  stageDependencies:\n    setup: []\n",
			nil, nil, []string{},
		),
		ginkgo.Entry("empty stageDependencies mapping",
			"- add: /\n  to: /app\n  stageDependencies: {}\n",
			nil, nil, nil,
		),
		ginkgo.Entry("all stages declared, masks and empty lists mixed",
			"- add: /\n  to: /app\n  stageDependencies:\n    install: []\n    beforeSetup: [\"src/**/*\"]\n    setup: []\n",
			[]string{}, []string{"src/**/*"}, []string{},
		),
		ginkgo.Entry("yaml keys in reverse stage order",
			"- add: /\n  to: /app\n  stageDependencies:\n    setup: []\n    beforeSetup: [\"src/**/*\"]\n    install: []\n",
			[]string{}, []string{"src/**/*"}, []string{},
		),
		ginkgo.Entry("scalar mask instead of a list",
			"- add: /\n  to: /app\n  stageDependencies:\n    install: package.json\n",
			[]string{"package.json"}, nil, nil,
		),
	)

	ginkgo.It("leaves stageDependencies unset when the block is absent", func() {
		gitLocals, err := parseStapelImageGitLocals(stapelImageWithGit("- add: /\n  to: /app\n"))
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(gitLocals).To(gomega.HaveLen(1))
		gomega.Expect(gitLocals[0].StageDependencies).To(gomega.BeNil())
	})

	ginkgo.It("converts each git mapping independently", func() {
		gitLocals, err := parseStapelImageGitLocals(stapelImageWithGit(
			"- add: /one\n  to: /app/one\n  stageDependencies:\n    install: []\n" +
				"- add: /two\n  to: /app/two\n  stageDependencies:\n    setup: [\"src/**/*\"]\n",
		))
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(gitLocals).To(gomega.HaveLen(2))

		gomega.Expect(gitLocals[0].StageDependencies.Install).To(gomega.Equal([]string{}))
		gomega.Expect(gitLocals[0].StageDependencies.BeforeSetup).To(gomega.BeNil())
		gomega.Expect(gitLocals[0].StageDependencies.Setup).To(gomega.BeNil())

		gomega.Expect(gitLocals[1].StageDependencies.Install).To(gomega.BeNil())
		gomega.Expect(gitLocals[1].StageDependencies.BeforeSetup).To(gomega.BeNil())
		gomega.Expect(gitLocals[1].StageDependencies.Setup).To(gomega.Equal([]string{"src/**/*"}))
	})

	ginkgo.DescribeTable("rejects invalid stage dependencies",
		func(gitYaml, expectedError string) {
			_, err := parseStapelImageGitLocals(stapelImageWithGit(gitYaml))
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedError)))
		},
		ginkgo.Entry("absolute install path",
			"- add: /\n  to: /app\n  stageDependencies:\n    install: [\"/package.json\"]\n",
			"`install: [PATH, ...]|PATH` should be relative paths!",
		),
		ginkgo.Entry("absolute beforeSetup path",
			"- add: /\n  to: /app\n  stageDependencies:\n    beforeSetup: [\"/src\"]\n",
			"`beforeSetup: [PATH, ...]|PATH` should be relative paths!",
		),
		ginkgo.Entry("absolute setup path",
			"- add: /\n  to: /app\n  stageDependencies:\n    setup: [\"/src\"]\n",
			"`setup: [PATH, ...]|PATH` should be relative paths!",
		),
		ginkgo.Entry("non-string mask",
			"- add: /\n  to: /app\n  stageDependencies:\n    install: [1]\n",
			"single string or array of strings expected",
		),
		ginkgo.Entry("mapping instead of masks",
			"- add: /\n  to: /app\n  stageDependencies:\n    setup: {a: b}\n",
			"single string or array of strings expected",
		),
	)
})
