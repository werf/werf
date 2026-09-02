package file_reader

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("chart .helmignore rules", func() {
	DescribeTable("matches a file against the chart rules",
		func(helmignore *string, relPath string, expected bool) {
			var data []byte
			if helmignore != nil {
				data = []byte(*helmignore)
			}

			rules, err := parseChartIgnoreRules(data, data != nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(matchChartIgnoreRules(rules.rules, relPath, false)).To(Equal(expected))
		},
		Entry("keeps a file when no rule matches", lo.ToPtr("templates/ignored.yaml\n"), "templates/kept.yaml", false),
		Entry("keeps every file when the rules are empty", lo.ToPtr(""), "templates/kept.yaml", false),

		Entry("matches an exact path", lo.ToPtr("templates/ignored.yaml\n"), "templates/ignored.yaml", true),
		Entry("matches a basename at any depth", lo.ToPtr("ignored.yaml\n"), "templates/sub/ignored.yaml", true),
		Entry("matches a glob", lo.ToPtr("*.txt\n"), "templates/notes.txt", true),
		Entry("does not match a path-anchored rule elsewhere", lo.ToPtr("templates/ignored.yaml\n"), "sub/templates/ignored.yaml", false),
		Entry("matches a root-anchored rule at the root", lo.ToPtr("/values.yaml\n"), "values.yaml", true),
		Entry("keeps a root-anchored rule's namesake deeper in the chart", lo.ToPtr("/values.yaml\n"), "templates/values.yaml", false),

		// A directory rule is matched on the directory itself, so it never matches a file.
		// The walk skips the whole subtree instead, which is covered by the LoadChartDir specs.
		Entry("does not match a file against a directory rule", lo.ToPtr("mydir/\n"), "mydir/inner.yaml", false),

		// Helm's default rules apply whether or not the chart carries a .helmignore.
		Entry("ignores dotfiles in templates without a .helmignore", nil, "templates/.gitkeep", true),
		Entry("ignores dotfiles in templates with a .helmignore", lo.ToPtr("unrelated.yaml\n"), "templates/.gitkeep", true),
		Entry("keeps dotfiles nested deeper than the default rule", nil, "templates/sub/.gitkeep", false),
		Entry("keeps dotfiles outside templates", nil, ".gitkeep", false),
		Entry("keeps regular templates without a .helmignore", nil, "templates/kept.yaml", false),
	)

	DescribeTable("matches a directory against the chart rules",
		func(helmignore, relPath string, expected bool) {
			rules, err := parseChartIgnoreRules([]byte(helmignore), true)
			Expect(err).NotTo(HaveOccurred())

			Expect(matchChartIgnoreRules(rules.rules, relPath, true)).To(Equal(expected))
		},
		Entry("matches a trailing-slash rule", "mydir/\n", "mydir", true),
		Entry("matches a rule written without a trailing slash", "logs\n", "logs", true),
		Entry("matches a nested directory by its path", "templates/sub/\n", "templates/sub", true),
		Entry("keeps a directory whose name only prefixes the rule", "mydir/\n", "mydir-other", false),
		Entry("keeps an unmatched directory", "mydir/\n", "templates", false),
	)

	DescribeTable("matches a file from a flat list, where no directory can be pruned by the walk",
		func(ctx SpecContext, helmignore, relPath string, expected bool) {
			rules, err := parseChartIgnoreRules([]byte(helmignore), true)
			Expect(err).NotTo(HaveOccurred())

			Expect(rules.IsFileIgnored(ctx, relPath)).To(Equal(expected))
		},
		Entry("matches the file itself", "templates/ignored.yaml\n", "templates/ignored.yaml", true),
		Entry("matches a file under a directory rule", "mydir/\n", "mydir/inner.yaml", true),
		Entry("matches a file deep under a directory rule", "mydir/\n", "mydir/deep/inner.yaml", true),
		Entry("matches a file under a directory rule written without a slash", "logs\n", "logs/inner.yaml", true),
		Entry("keeps a file whose directory only prefixes the rule", "mydir/\n", "mydir-other/inner.yaml", false),
		Entry("keeps an unmatched file", "mydir/\n", "templates/kept.yaml", false),
	)

	// The zero value is reachable through the exported API, where a nil rule set used to panic.
	It("excludes nothing when the rule set is the zero value", func(ctx SpecContext) {
		Expect(ChartIgnoreRules{}.IsFileIgnored(ctx, "templates/kept.yaml")).To(BeFalse())
	})

	It("reads the basename from the file info, so a basename rule keeps working if helm starts using it", func() {
		Expect(chartIgnoreFileInfo{path: "templates/sub/ignored.yaml"}.Name()).To(Equal("ignored.yaml"))
	})

	It("fails on a rule helm cannot compile", func() {
		_, err := parseChartIgnoreRules([]byte("templates/**/ignored.yaml\n"), true)
		Expect(err).To(MatchError(ContainSubstring("double-star")))
	})
})

var _ = Describe("ChartIgnoreRules presence", func() {
	// An empty .helmignore excludes nothing yet is still a .helmignore, so presence has to be
	// carried alongside the rules instead of being inferred from them.
	It("reports an empty .helmignore as present", func() {
		rules, err := parseChartIgnoreRules([]byte(""), true)
		Expect(err).NotTo(HaveOccurred())
		Expect(rules.HasIgnoreFile()).To(BeTrue())
	})

	It("reports a chart without a .helmignore as having none", func() {
		rules, err := parseChartIgnoreRules(nil, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(rules.HasIgnoreFile()).To(BeFalse())
	})

	It("reports helm's default rule set as having no .helmignore", func() {
		Expect(DefaultChartIgnoreRules().HasIgnoreFile()).To(BeFalse())
	})
})
