package common

import (
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/docker_registry"
)

var _ = ginkgo.Describe("Repository delete credentials", func() {
	ginkgo.DescribeTable("registers credentials only when requested", func(ctx ginkgo.SpecContext, name string, deleteCredentials bool) {
		cmd := &cobra.Command{}
		repo := NewRepoData(name, RepoDataOptions{})
		suffixes := []string{"docker-hub-username", "docker-hub-password", "docker-hub-token", "github-token", "harbor-username", "harbor-password", "quay-token"}
		for _, suffix := range suffixes {
			env := "WERF_" + strings.ToUpper(strings.ReplaceAll(name+"-"+suffix, "-", "_"))
			ginkgo.GinkgoT().Setenv(env, "env-"+suffix)
		}
		repo.SetupCmd(cmd)
		if deleteCredentials {
			repo.SetupDeleteCredentials(ctx, cmd)
		}
		gomega.Expect(cmd.Flags().Lookup(name)).NotTo(gomega.BeNil())
		gomega.Expect(cmd.Flags().Lookup(name + "-container-registry")).NotTo(gomega.BeNil())
		for _, suffix := range suffixes {
			flag := cmd.Flags().Lookup(name + "-" + suffix)
			if deleteCredentials {
				gomega.Expect(flag).NotTo(gomega.BeNil())
				gomega.Expect(flag.DefValue).To(gomega.Equal("env-" + suffix))
				gomega.Expect(cmd.Flags().Set(name+"-"+suffix, "flag-"+suffix)).To(gomega.Succeed())
			} else {
				gomega.Expect(flag).To(gomega.BeNil())
			}
		}
		expected := docker_registry.DockerRegistryOptions{
			InsecureRegistry:      true,
			SkipTlsVerifyRegistry: true,
			InsecureRegistryHosts: []string{"registry.example"},
		}
		if deleteCredentials {
			expected.DockerHubUsername = "flag-docker-hub-username"
			expected.DockerHubPassword = "flag-docker-hub-password"
			expected.DockerHubToken = "flag-docker-hub-token"
			expected.GitHubToken = "flag-github-token"
			expected.HarborUsername = "flag-harbor-username"
			expected.HarborPassword = "flag-harbor-password"
			expected.QuayToken = "flag-quay-token"
		}
		gomega.Expect(repo.GetDockerRegistryOptions(true, true, []string{"registry.example"})).To(gomega.Equal(expected))
	},
		ginkgo.Entry("ordinary repo", "repo", false),
		ginkgo.Entry("ordinary final repo", "final-repo", false),
		ginkgo.Entry("repo deletion", "repo", true),
		ginkgo.Entry("final repo deletion", "final-repo", true),
	)
})
