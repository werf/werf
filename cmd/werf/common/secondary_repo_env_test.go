package common

import (
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = ginkgo.Describe("Secondary repository environment", func() {
	ginkgo.DescribeTable("combines environment values and repeated flags", func(repo string, expected []string) {
		for _, entry := range os.Environ() {
			key, _, _ := strings.Cut(entry, "=")
			if strings.HasPrefix(key, "WERF_SECONDARY_REPO_") {
				ginkgo.GinkgoT().Setenv(key, "")
				gomega.Expect(os.Unsetenv(key)).To(gomega.Succeed())
			}
		}
		ginkgo.GinkgoT().Setenv("WERF_SECONDARY_REPO", repo)
		ginkgo.GinkgoT().Setenv("WERF_SECONDARY_REPO_2", "example.com/second")
		ginkgo.GinkgoT().Setenv("WERF_SECONDARY_REPO_1", "example.com/first")
		data := &CmdData{}
		command := &cobra.Command{}
		SetupSecondaryStagesStorageOptions(data, command)
		gomega.Expect(command.ParseFlags([]string{"--secondary-repo=example.com/cli-one", "--secondary-repo=example.com/cli-two"})).To(gomega.Succeed())
		gomega.Expect(GetSecondaryStagesStorage(data)).To(gomega.Equal(expected))
	},
		ginkgo.Entry("bare variable precedes suffixed variables and flags", "example.com/bare", []string{"example.com/bare", "example.com/first", "example.com/second", "example.com/cli-one", "example.com/cli-two"}),
		ginkgo.Entry("empty bare variable adds no repository", "", []string{"example.com/first", "example.com/second", "example.com/cli-one", "example.com/cli-two"}),
	)
})
