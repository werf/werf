package image

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/build/stage"
)

const stageDependenciesFixtureYaml = `project: test
configVersion: 1
---
image: app
from: alpine:3.20
git:
- add: /src
  to: /app
%s
shell:
  beforeInstall: ["echo beforeInstall"]
  install: ["echo install"]
  beforeSetup: ["echo beforeSetup"]
  setup: ["echo setup"]
`

var _ = ginkgo.DescribeTable("git stageDependencies config to git mapping wiring",
	func(stageDependenciesBlock string, expect func(map[stage.StageName][]string)) {
		ctx := context.Background()
		projectDir := newProjectRepo(ctx, stageDependenciesFixture(stageDependenciesBlock))

		gitMappings := gitMappingsOf(ctx, projectDir)
		gomega.Expect(gitMappings).To(gomega.HaveLen(1))

		expect(gitMappings[0].StagesDependencies)
	},
	ginkgo.Entry("absent stageDependencies declares nothing", "", func(deps map[stage.StageName][]string) {
		gomega.Expect(deps).To(gomega.BeNil())
	}),
	ginkgo.Entry("empty stageDependencies declares nothing", "  stageDependencies: {}", func(deps map[stage.StageName][]string) {
		gomega.Expect(deps).To(gomega.BeEmpty())
	}),
	ginkgo.Entry("only the declared stage is a key", `  stageDependencies:
    beforeSetup: ["*"]`, func(deps map[stage.StageName][]string) {
		gomega.Expect(deps).To(gomega.HaveLen(1))
		gomega.Expect(deps).To(gomega.HaveKeyWithValue(stage.BeforeSetup, []string{"*"}))
	}),
	ginkgo.Entry("an explicitly empty list stays declared", `  stageDependencies:
    install: []
    setup: ["cfg/**/*"]`, func(deps map[stage.StageName][]string) {
		gomega.Expect(deps).To(gomega.HaveLen(2))
		gomega.Expect(deps).To(gomega.HaveKeyWithValue(stage.Install, []string{}))
		gomega.Expect(deps).To(gomega.HaveKeyWithValue(stage.Setup, []string{"cfg/**/*"}))
	}),
)

var _ = ginkgo.DescribeTable("git stageDependencies effective checksums",
	func(stageDependenciesBlock string, edits map[string]string, expectedChanged []string) {
		ctx := context.Background()
		projectDir := newProjectRepo(ctx, stageDependenciesFixture(stageDependenciesBlock))

		gomega.Expect(changedChecksums(ctx, projectDir, edits)).To(gomega.ConsistOf(expectedChanged))
	},
	ginkgo.Entry("absent stageDependencies keeps every stage on all files",
		"",
		map[string]string{"src/app/main.go": "package main // changed\n"},
		[]string{"install[0]", "beforeSetup[0]", "setup[0]"},
	),
	ginkgo.Entry("empty stageDependencies keeps every stage on all files",
		"  stageDependencies: {}",
		map[string]string{"src/app/main.go": "package main // changed\n"},
		[]string{"install[0]", "beforeSetup[0]", "setup[0]"},
	),
	ginkgo.Entry("stage declared last: earlier stage gets nothing, later gets all files",
		`  stageDependencies:
    beforeSetup: ["*"]`,
		map[string]string{"src/root.txt": "root changed\n"},
		[]string{"beforeSetup[0]", "setup[0]"},
	),
	ginkgo.Entry("stage declared last: a file outside the git mapping changes nothing",
		`  stageDependencies:
    beforeSetup: ["*"]`,
		map[string]string{"outside.txt": "outside changed\n"},
		nil,
	),
	ginkgo.Entry("stage declared first: later stages get all files",
		`  stageDependencies:
    install: ["app/**/*"]`,
		map[string]string{"src/app/main.go": "package main // changed\n"},
		[]string{"install[0]", "beforeSetup[0]", "setup[0]"},
	),
	ginkgo.Entry("stage declared first: its own mask does not match a sibling file",
		`  stageDependencies:
    install: ["app/**/*"]`,
		map[string]string{"src/root.txt": "root changed\n"},
		[]string{"beforeSetup[0]", "setup[0]"},
	),
	ginkgo.Entry("explicitly empty list and a later declaration pin both earlier stages",
		`  stageDependencies:
    install: []
    setup: ["cfg/**/*"]`,
		map[string]string{"src/cfg/values.yaml": "key: changed\n", "src/root.txt": "root changed\n"},
		[]string{"setup[0]"},
	),
	ginkgo.Entry("all stages declared, yaml keys reordered: only the matching mask reacts",
		`  stageDependencies:
    setup: ["cfg/**/*"]
    beforeSetup: ["root.txt"]
    install: ["app/**/*"]`,
		map[string]string{"src/app/main.go": "package main // changed\n"},
		[]string{"install[0]"},
	),
	ginkgo.Entry("all stages declared, yaml keys reordered: unmatched file changes nothing",
		`  stageDependencies:
    setup: ["cfg/**/*"]
    beforeSetup: ["root.txt"]
    install: ["app/**/*"]`,
		map[string]string{"outside.txt": "outside changed\n"},
		nil,
	),
)

var _ = ginkgo.Describe("git stageDependencies of independent git mappings", func() {
	const werfYaml = `project: test
configVersion: 1
---
image: app
from: alpine:3.20
git:
- add: /src
  to: /app
  stageDependencies:
    install: ["**/*"]
- add: /cfg
  to: /cfg
  stageDependencies:
    setup: ["**/*"]
shell:
  install: ["echo install"]
  beforeSetup: ["echo beforeSetup"]
  setup: ["echo setup"]
`

	newRepo := func(ctx context.Context) string {
		return newProjectRepo(ctx, map[string]string{
			"werf.yaml":       werfYaml,
			"src/app/main.go": "package main\n",
			"cfg/values.yaml": "key: value\n",
		})
	}

	ginkgo.It("resolves each git mapping on its own declarations", func() {
		ctx := context.Background()
		projectDir := newRepo(ctx)

		gitMappings := gitMappingsOf(ctx, projectDir)
		gomega.Expect(gitMappings).To(gomega.HaveLen(2))
		gomega.Expect(gitMappings[0].StagesDependencies).To(gomega.HaveLen(1))
		gomega.Expect(gitMappings[0].StagesDependencies).To(gomega.HaveKeyWithValue(stage.Install, []string{"**/*"}))
		gomega.Expect(gitMappings[1].StagesDependencies).To(gomega.HaveLen(1))
		gomega.Expect(gitMappings[1].StagesDependencies).To(gomega.HaveKeyWithValue(stage.Setup, []string{"**/*"}))
	})

	ginkgo.It("changes only the stages of the git mapping owning the changed file", func() {
		ctx := context.Background()
		projectDir := newRepo(ctx)

		gomega.Expect(changedChecksums(ctx, projectDir, map[string]string{"src/app/main.go": "package main // changed\n"})).
			To(gomega.ConsistOf("install[0]", "beforeSetup[0]", "setup[0]"))
		gomega.Expect(changedChecksums(ctx, projectDir, map[string]string{"cfg/values.yaml": "key: changed\n"})).
			To(gomega.ConsistOf("setup[1]"))
	})
})

var _ = ginkgo.Describe("git stageDependencies within the git mapping path scope", func() {
	const werfYaml = `project: test
configVersion: 1
---
image: app
from: alpine:3.20
git:
- add: /src
  to: /app
  includePaths: ["keep"]
  excludePaths: ["keep/skip"]
  stageDependencies:
    install: ["**/*"]
shell:
  install: ["echo install"]
  beforeSetup: ["echo beforeSetup"]
  setup: ["echo setup"]
`

	newRepo := func(ctx context.Context) string {
		return newProjectRepo(ctx, map[string]string{
			"werf.yaml":               werfYaml,
			"src/keep/kept.txt":       "kept\n",
			"src/keep/skip/other.txt": "excluded\n",
			"src/dropped.txt":         "not included\n",
			"outside.txt":             "outside add\n",
		})
	}

	ginkgo.It("reacts to a file inside include paths", func() {
		ctx := context.Background()
		gomega.Expect(changedChecksums(ctx, newRepo(ctx), map[string]string{"src/keep/kept.txt": "kept changed\n"})).
			To(gomega.ConsistOf("install[0]", "beforeSetup[0]", "setup[0]"))
	})

	ginkgo.It("ignores files cut off by excludePaths, includePaths and add", func() {
		ctx := context.Background()
		projectDir := newRepo(ctx)

		gomega.Expect(changedChecksums(ctx, projectDir, map[string]string{"src/keep/skip/other.txt": "excluded changed\n"})).To(gomega.BeEmpty())
		gomega.Expect(changedChecksums(ctx, projectDir, map[string]string{"src/dropped.txt": "not included changed\n"})).To(gomega.BeEmpty())
		gomega.Expect(changedChecksums(ctx, projectDir, map[string]string{"outside.txt": "outside changed\n"})).To(gomega.BeEmpty())
	})
})

var _ = ginkgo.Describe("git stageDependencies without stage instructions", func() {
	werfYaml := func(stageDependenciesBlock string) string {
		return fmt.Sprintf(`project: test
configVersion: 1
---
image: app
from: alpine:3.20
git:
- add: /src
  to: /app
%s
shell:
  beforeInstall: ["echo beforeInstall"]
`, stageDependenciesBlock)
	}

	newRepo := func(ctx context.Context, stageDependenciesBlock string) string {
		return newProjectRepo(ctx, map[string]string{
			"werf.yaml":       werfYaml(stageDependenciesBlock),
			"src/app/main.go": "package main\n",
		})
	}

	ginkgo.It("still rejects a non-empty declaration", func() {
		ctx := context.Background()
		projectDir := newRepo(ctx, `  stageDependencies:
    install: ["**/*"]`)

		_, err := stapelImage(ctx, projectDir)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("git.stageDependencies.install is defined, but no install instructions are provided")))
	})

	ginkgo.It("accepts an explicitly empty declaration and generates no user stages for it", func() {
		ctx := context.Background()
		projectDir := newRepo(ctx, `  stageDependencies:
    install: []`)

		image, err := stapelImage(ctx, projectDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageNames(image)).To(gomega.ContainElement(stage.BeforeInstall))
		gomega.Expect(stageNames(image)).NotTo(gomega.ContainElement(stage.Install))
	})

	ginkgo.It("accepts absent declarations and generates no user stages for them", func() {
		ctx := context.Background()
		projectDir := newRepo(ctx, "")

		image, err := stapelImage(ctx, projectDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(stageNames(image)).To(gomega.ContainElement(stage.BeforeInstall))
		gomega.Expect(stageNames(image)).NotTo(gomega.ContainElement(stage.Install))
		gomega.Expect(stageNames(image)).NotTo(gomega.ContainElement(stage.BeforeSetup))
		gomega.Expect(stageNames(image)).NotTo(gomega.ContainElement(stage.Setup))
	})

	ginkgo.It("keeps the beforeInstall stage independent of git file changes", func() {
		ctx := context.Background()
		projectDir := newRepo(ctx, `  stageDependencies:
    install: []`)

		before, err := beforeInstallDependencies(ctx, projectDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		commitFiles(ctx, projectDir, map[string]string{"src/app/main.go": "package main // changed\n"})

		after, err := beforeInstallDependencies(ctx, projectDir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(after).To(gomega.Equal(before))
	})
})
