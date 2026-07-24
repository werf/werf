package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
)

type entry struct {
	repository string
	expectedID string
}

var _ = DescribeTable("parsing git repository ID", func(e entry) {
	Expect(getRepositoryID(e.repository)).Should(Equal(e.expectedID))
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

var _ = Describe("gitRemoteRepoCacheKey", func() {
	It("distinguishes refs and auth contexts", func() {
		url := "https://example.com/repo.git"
		base := gitRemoteRepoCacheKey(url, "", "v1", "", nil)
		Expect(gitRemoteRepoCacheKey(url, "", "v1", "", nil)).To(Equal(base))
		Expect(gitRemoteRepoCacheKey(url, "", "v2", "", nil)).NotTo(Equal(base))
		Expect(gitRemoteRepoCacheKey(url, "", "", "0123456789012345678901234567890123456789", nil)).NotTo(Equal(base))
		Expect(gitRemoteRepoCacheKey(url, "", "v1", "", &git_repo.BasicAuthCredentials{
			Username: "user",
			Password: git_repo.PasswordSource{Env: "PASSWORD"},
		})).NotTo(Equal(base))
	})
})
