package checker

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("checker", func() {
	Describe("parseResult", func() {
		DescribeTable("parses output correctly",
			func(out, fileName string, index, total int, matcher types.GomegaMatcher) {
				err := parseResult(context.Background(), out, fileName, index, total)
				Expect(err).To(matcher)
			},
			Entry("no errors no warnings",
				"файл корректный\n", "valid.json", 1, 1,
				Succeed()),
			Entry("empty output",
				"", "empty.json", 1, 1,
				Succeed()),
			Entry("errors only",
				"ERROR: missing bomFormat\nERROR: missing specVersion\n", "bad.json", 1, 3,
				MatchError(ContainSubstring("validation failed for bad.json"))),
			Entry("warnings only",
				"WARNING: vcs url not found for pkg1\nWARNING: vcs url not found for pkg2\n", "warn.json", 2, 3,
				MatchError(ContainSubstring("validation failed for warn.json"))),
			Entry("errors and warnings",
				"ERROR: bad field\nWARNING: vcs issue\n", "mixed.json", 1, 1,
				MatchError(ContainSubstring("validation failed for mixed.json"))),
			Entry("non-prefixed output only",
				"some random output\nanother line\n", "random.json", 1, 2,
				Succeed()),
			Entry("errors mixed with non-prefixed lines",
				"starting check\nERROR: bad field\ndone\n", "report.json", 3, 5,
				MatchError(ContainSubstring("validation failed for report.json"))),
			Entry("error details included in message",
				"ERROR: missing bomFormat\nERROR: missing specVersion\n", "bad.json", 1, 1,
				MatchError(And(
					ContainSubstring("ERROR: missing bomFormat"),
					ContainSubstring("ERROR: missing specVersion"),
				))),
			Entry("warning details included in message",
				"WARNING: vcs url not found\n", "warn.json", 1, 1,
				MatchError(ContainSubstring("WARNING: vcs url not found"))),
		)
	})

	Describe("buildDockerArgs", func() {
		DescribeTable("builds correct docker arguments",
			func(path string, isprasFormat IsprasFormat, checkVCS bool, want []string) {
				got, err := buildDockerArgs(path, isprasFormat, checkVCS)
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(want))
			},
			Entry("oss without check-vcs",
				"/tmp/sbom.json", IsprasFormatOSS, false,
				[]string{
					"--rm",
					"-v", "/tmp/sbom.json:/sbom/input.json:ro",
					Image,
					"--format", "oss", "--errors", "0", "/sbom/input.json",
				}),
			Entry("oss with check-vcs",
				"/tmp/sbom.json", IsprasFormatOSS, true,
				[]string{
					"--rm",
					"-v", "/tmp/sbom.json:/sbom/input.json:ro",
					Image,
					"--format", "oss", "--errors", "0", "--check-vcs", "/sbom/input.json",
				}),
			Entry("container format",
				"/tmp/sbom.json", IsprasFormatContainer, false,
				[]string{
					"--rm",
					"-v", "/tmp/sbom.json:/sbom/input.json:ro",
					Image,
					"--format", "container", "--errors", "0", "/sbom/input.json",
				}),
			Entry("container with check-vcs",
				"/tmp/sbom.json", IsprasFormatContainer, true,
				[]string{
					"--rm",
					"-v", "/tmp/sbom.json:/sbom/input.json:ro",
					Image,
					"--format", "container", "--errors", "0", "--check-vcs", "/sbom/input.json",
				}),
		)
	})

	Describe("extractPrefixedLines", func() {
		DescribeTable("extracts lines with given prefix",
			func(text, prefix string, want []string) {
				Expect(extractPrefixedLines(text, prefix)).To(Equal(want))
			},
			Entry("no matching lines",
				"some output\nanother line\n", "ERROR:",
				[]string(nil)),
			Entry("single error line",
				"ERROR: bad field\n", "ERROR:",
				[]string{"ERROR: bad field"}),
			Entry("multiple error lines",
				"ERROR: first\nok\nERROR: second\n", "ERROR:",
				[]string{"ERROR: first", "ERROR: second"}),
			Entry("warning lines",
				"WARNING: vcs not found\nWARNING: another\n", "WARNING:",
				[]string{"WARNING: vcs not found", "WARNING: another"}),
			Entry("trims whitespace before matching",
				"  ERROR: indented\n\tERROR: tabbed\n", "ERROR:",
				[]string{"ERROR: indented", "ERROR: tabbed"}),
			Entry("empty input",
				"", "ERROR:",
				[]string(nil)),
		)
	})
})
