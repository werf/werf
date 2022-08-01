package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type entry struct {
	repository string
	expectedID string
}

var _ = DescribeTable("parsing git repository ID", func(e entry) {
	Ω(getRepositoryID(e.repository)).Should(Equal(e.expectedID))
},
	Entry("git", entry{
		"git@github.com:company/name.git",
		"company/name",
	}),
	Entry("git without ending", entry{
		"git@github.com:company/name",
		"company/name",
	}),
	Entry("https", entry{
		"https://github.com/company/name.git",
		"company/name",
	}),
	Entry("https with credentials", entry{
		"https://username:password@github.com/company/name.git",
		"company/name",
	}),
	Entry("file", entry{
		"file:///path/workspace/name.git/",
		"workspace/name",
	}),
	Entry("relative", entry{
		"../name",
		"../name",
	}))
