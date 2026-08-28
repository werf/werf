package file_reader

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("chart .helmignore rules", func() {
	DescribeTable("matches a file against the chart rules",
		func(helmignore *string, relPath string, expected bool) {
			var data []byte
			if helmignore != nil {
				data = []byte(*helmignore)
			}

			rules, err := parseChartIgnoreRules(data)
			Expect(err).NotTo(HaveOccurred())

			Expect(matchChartIgnoreRules(rules, relPath, false)).To(Equal(expected))
		},
		Entry("keeps a file when no rule matches", helmignore("templates/ignored.yaml\n"), "templates/kept.yaml", false),
		Entry("keeps every file when the rules are empty", helmignore(""), "templates/kept.yaml", false),

		Entry("matches an exact path", helmignore("templates/ignored.yaml\n"), "templates/ignored.yaml", true),
		Entry("matches a basename at any depth", helmignore("ignored.yaml\n"), "templates/sub/ignored.yaml", true),
		Entry("matches a glob", helmignore("*.txt\n"), "templates/notes.txt", true),
		Entry("does not match a path-anchored rule elsewhere", helmignore("templates/ignored.yaml\n"), "sub/templates/ignored.yaml", false),
		Entry("matches a root-anchored rule at the root", helmignore("/values.yaml\n"), "values.yaml", true),
		Entry("keeps a root-anchored rule's namesake deeper in the chart", helmignore("/values.yaml\n"), "templates/values.yaml", false),

		// A directory rule is matched on the directory itself, so it never matches a file.
		// The walk skips the whole subtree instead, which is covered by the LoadChartDir specs.
		Entry("does not match a file against a directory rule", helmignore("mydir/\n"), "mydir/inner.yaml", false),

		// Helm's default rules apply whether or not the chart carries a .helmignore.
		Entry("ignores dotfiles in templates without a .helmignore", nil, "templates/.gitkeep", true),
		Entry("ignores dotfiles in templates with a .helmignore", helmignore("unrelated.yaml\n"), "templates/.gitkeep", true),
		Entry("keeps dotfiles nested deeper than the default rule", nil, "templates/sub/.gitkeep", false),
		Entry("keeps dotfiles outside templates", nil, ".gitkeep", false),
		Entry("keeps regular templates without a .helmignore", nil, "templates/kept.yaml", false),
	)

	DescribeTable("matches a directory against the chart rules",
		func(helmignore, relPath string, expected bool) {
			rules, err := parseChartIgnoreRules([]byte(helmignore))
			Expect(err).NotTo(HaveOccurred())

			Expect(matchChartIgnoreRules(rules, relPath, true)).To(Equal(expected))
		},
		Entry("matches a trailing-slash rule", "mydir/\n", "mydir", true),
		Entry("matches a rule written without a trailing slash", "logs\n", "logs", true),
		Entry("matches a nested directory by its path", "templates/sub/\n", "templates/sub", true),
		Entry("keeps a directory whose name only prefixes the rule", "mydir/\n", "mydir-other", false),
		Entry("keeps an unmatched directory", "mydir/\n", "templates", false),
	)

	It("fails on a rule helm cannot compile", func() {
		_, err := parseChartIgnoreRules([]byte("templates/**/ignored.yaml\n"))
		Expect(err).To(MatchError(ContainSubstring("double-star")))
	})
})

func helmignore(data string) *string {
	return &data
}
