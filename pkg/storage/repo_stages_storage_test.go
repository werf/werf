package storage

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RepoStagesStorage", func() {
	DescribeTable("makeRepoSbomImageRecord",
		func(repoAddress, imageName, expected string) {
			Expect(makeRepoSbomImageRecord(repoAddress, imageName)).To(Equal(expected))
		},
		Entry(
			"should work with remote image name",
			"localhost:35907/werf-test-none-169721-57a20c5b",
			"localhost:35907/werf-test-none-169721-57a20c5b:f21387755c3bda895a39b6517300e26f3bab0a23ecf800c35d76fed6-1748283024291",
			"localhost:35907/werf-test-none-169721-57a20c5b:f21387755c3bda895a39b6517300e26f3bab0a23ecf800c35d76fed6-1748283024291",
		),
	)
})
