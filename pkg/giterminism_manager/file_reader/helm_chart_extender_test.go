package file_reader

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("chart .helmignore rules", func() {
	DescribeTable("matches the path against the chart rules",
		func(helmignore *string, relPath string, expected bool) {
			var data []byte
			if helmignore != nil {
				data = []byte(*helmignore)
			}

			rules, err := parseChartIgnoreRules(data)
			Expect(err).NotTo(HaveOccurred())

			Expect(matchChartIgnoreRules(rules, relPath)).To(Equal(expected))
		},
		Entry("keeps a file when no rule matches", helmignore("templates/ignored.yaml\n"), "templates/kept.yaml", false),
		Entry("keeps every file when the rules are empty", helmignore(""), "templates/kept.yaml", false),

		Entry("matches an exact path", helmignore("templates/ignored.yaml\n"), "templates/ignored.yaml", true),
		Entry("matches a basename at any depth", helmignore("ignored.yaml\n"), "templates/sub/ignored.yaml", true),
		Entry("matches a glob", helmignore("*.txt\n"), "templates/notes.txt", true),
		Entry("does not match a path-anchored rule elsewhere", helmignore("templates/ignored.yaml\n"), "sub/templates/ignored.yaml", false),
		Entry("matches a root-anchored rule at the root", helmignore("/values.yaml\n"), "values.yaml", true),

		Entry("matches a file under a directory rule", helmignore("mydir/\n"), "mydir/inner.yaml", true),
		Entry("matches a file deep under a directory rule", helmignore("mydir/\n"), "mydir/deep/inner.yaml", true),
		Entry("matches a directory named without a trailing slash", helmignore("logs\n"), "logs/output.txt", true),
		Entry("keeps a file whose name only prefixes a directory rule", helmignore("mydir/\n"), "mydir-other/inner.yaml", false),

		// Helm treats "!" as inverting the whole rule set rather than re-including a file,
		// so a negated rule ignores everything it does not match. Documented as unsupported.
		Entry("ignores a non-matching path when a negated rule is present", helmignore("!templates/kept.yaml\n"), "templates/other.yaml", true),
		Entry("ignores even the negated path itself", helmignore("!templates/kept.yaml\n"), "templates/kept.yaml", true),

		// Helm's default rules apply whether or not the chart carries a .helmignore.
		Entry("ignores dotfiles in templates without a .helmignore", nil, "templates/.gitkeep", true),
		Entry("ignores dotfiles in templates with a .helmignore", helmignore("unrelated.yaml\n"), "templates/.gitkeep", true),
		Entry("keeps dotfiles nested deeper than the default rule", nil, "templates/sub/.gitkeep", false),
		Entry("keeps dotfiles outside templates", nil, ".gitkeep", false),
		Entry("keeps regular templates without a .helmignore", nil, "templates/kept.yaml", false),
	)

	It("fails on a rule helm cannot compile", func() {
		_, err := parseChartIgnoreRules([]byte("templates/**/ignored.yaml\n"))
		Expect(err).To(MatchError(ContainSubstring("double-star")))
	})
})

func helmignore(data string) *string {
	return &data
}
