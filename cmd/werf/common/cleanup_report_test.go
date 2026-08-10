package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

type cleanupReportPathEntry struct {
	path     *string
	expected string
	errMsg   string
}

var _ = DescribeTable("GetCleanupReportPath",
	func(entry cleanupReportPathEntry) {
		path, err := GetCleanupReportPath(&CmdData{CleanupReportPath: entry.path})

		if entry.errMsg != "" {
			Expect(err).To(MatchError(entry.errMsg))
			return
		}

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(entry.expected))
	},
	Entry("unset", cleanupReportPathEntry{path: nil, expected: DefaultCleanupReportPathJSON}),
	Entry("empty", cleanupReportPathEntry{path: lo.ToPtr(""), expected: DefaultCleanupReportPathJSON}),
	Entry("json", cleanupReportPathEntry{path: lo.ToPtr("reports/cleanup.json"), expected: "reports/cleanup.json"}),
	Entry("no extension", cleanupReportPathEntry{path: lo.ToPtr("reports/cleanup"), expected: "reports/cleanup.json"}),
	Entry("other extension", cleanupReportPathEntry{
		path:   lo.ToPtr("reports/cleanup.yaml"),
		errMsg: `invalid --cleanup-report-path "reports/cleanup.yaml": extension must be either .json or unspecified`,
	}),
)
