package true_git

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseRawDiffPaths", func() {
	DescribeTable("parsing git diff --raw -z --no-renames output",
		func(out string, expected []ChangedPath) {
			changedPaths, err := parseRawDiffPaths(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(changedPaths).To(Equal(expected))
		},
		Entry("empty output", "", nil),
		Entry("modified file",
			":100644 100644 bcd1234567890123456789012345678901234567 0123456789012345678901234567890123456789 M\x00dir/file.go\x00",
			[]ChangedPath{{Path: "dir/file.go"}},
		),
		Entry("added and deleted files",
			":000000 100644 0000000000000000000000000000000000000000 1111111111111111111111111111111111111111 A\x00new.txt\x00"+
				":100755 000000 2222222222222222222222222222222222222222 0000000000000000000000000000000000000000 D\x00old.sh\x00",
			[]ChangedPath{{Path: "new.txt"}, {Path: "old.sh"}},
		),
		Entry("changed submodule pin",
			":160000 160000 3333333333333333333333333333333333333333 4444444444444444444444444444444444444444 M\x00sub\x00",
			[]ChangedPath{{Path: "sub", IsSubmodule: true}},
		),
		Entry("path converted to submodule",
			":100644 160000 5555555555555555555555555555555555555555 6666666666666666666666666666666666666666 T\x00sub\x00",
			[]ChangedPath{{Path: "sub", IsSubmodule: true}},
		),
		Entry("path with spaces and unicode",
			":100644 100644 bcd1234567890123456789012345678901234567 0123456789012345678901234567890123456789 M\x00dir with spaces/файл.txt\x00",
			[]ChangedPath{{Path: "dir with spaces/файл.txt"}},
		),
	)

	It("fails on malformed entry", func() {
		_, err := parseRawDiffPaths("garbage\x00path\x00")
		Expect(err).To(HaveOccurred())
	})
})
