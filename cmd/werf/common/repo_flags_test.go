package common

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("RepoFlags", func() {
	It("should register all expected flags", func() {
		cmd := &cobra.Command{Use: "test"}
		cmdData := &CmdData{}
		SetupRepoOptions(cmdData, cmd, RepoDataOptions{})

		for _, flagName := range []string{"images-repo", "meta-repo", "cache-from", "cache-to"} {
			Expect(cmd.Flags().Lookup(flagName)).ToNot(BeNil(), "flag --%s should be registered", flagName)
		}
	})

	Context("deprecation alias resolution", func() {
		It("should map --final-repo to --images-repo when images-repo is empty", func() {
			cmdData := &CmdData{}

			finalAddr := "registry.example.com/final"
			cmdData.deprecatedFinalRepoAddr = &finalAddr

			cmdData.ImagesRepo = &[]string{}

			err := resolveDeprecatedFlags(context.Background(), cmdData)
			Expect(err).ToNot(HaveOccurred())
			Expect(*cmdData.ImagesRepo).To(Equal([]string{"registry.example.com/final"}))
		})

		It("should error when --final-repo and --images-repo both set", func() {
			cmdData := &CmdData{}

			finalAddr := "registry.example.com/final"
			cmdData.deprecatedFinalRepoAddr = &finalAddr

			cmdData.ImagesRepo = &[]string{"registry.example.com/images"}

			err := resolveDeprecatedFlags(context.Background(), cmdData)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("--final-repo and --images-repo cannot be used together"))
		})
	})

	DescribeTable("conflict detection",
		func(repo, imagesRepo, metaRepo string, cacheFrom, cacheTo []string, expectErr bool) {
			cmdData := &CmdData{
				Repo:       NewRepoData("repo", RepoDataOptions{}),
				ImagesRepo: &[]string{},
				MetaRepo:   NewRepoData("meta-repo", RepoDataOptions{OnlyAddress: true}),
				CacheFrom:  &[]string{},
				CacheTo:    &[]string{},
			}
			cmdData.Repo.Address = &repo
			if imagesRepo != "" {
				cmdData.ImagesRepo = &[]string{imagesRepo}
			}
			cmdData.MetaRepo.Address = &metaRepo
			if cacheFrom != nil {
				cmdData.CacheFrom = &cacheFrom
			}
			if cacheTo != nil {
				cmdData.CacheTo = &cacheTo
			}

			err := validateRepoFlags(cmdData)
			if expectErr {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--repo cannot be combined with"))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("--repo alone is valid",
			"registry.example.com/repo", "", "", nil, nil, false),
		Entry("--repo + --images-repo conflicts",
			"registry.example.com/repo", "registry.example.com/images", "", nil, nil, true),
		Entry("granular flags without --repo is valid",
			"", "registry.example.com/images", "registry.example.com/meta", []string{"registry.example.com/cache"}, nil, false),
	)

	It("should detect conflict when WERF_REPO env var is set alongside --images-repo", func() {
		cmdData := &CmdData{
			Repo:       NewRepoData("repo", RepoDataOptions{}),
			ImagesRepo: &[]string{"registry.example.com/images"},
		}
		repoAddr := ""
		cmdData.Repo.Address = &repoAddr

		GinkgoT().Setenv("WERF_REPO", "registry.example.com/repo")

		err := validateRepoFlags(cmdData)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("--repo cannot be combined with"))
	})
})

var _ = Describe("StorageAssembly", func() {
	It("should preserve HostPurge field in NewStorageManagerConfig", func() {
		c := NewStorageManagerConfig{HostPurge: true}
		Expect(c.HostPurge).To(BeTrue())
	})

	DescribeTable("granular mode detection",
		func(imagesRepo, metaRepo string, cacheFrom, cacheTo []string, expected bool) {
			cmdData := &CmdData{
				ImagesRepo: &[]string{},
				MetaRepo:   NewRepoData("meta-repo", RepoDataOptions{OnlyAddress: true}),
				CacheFrom:  &[]string{},
				CacheTo:    &[]string{},
			}
			if imagesRepo != "" {
				cmdData.ImagesRepo = &[]string{imagesRepo}
			}
			cmdData.MetaRepo.Address = &metaRepo
			if cacheFrom != nil {
				cmdData.CacheFrom = &cacheFrom
			}
			if cacheTo != nil {
				cmdData.CacheTo = &cacheTo
			}

			Expect(getGranularMode(cmdData)).To(Equal(expected))
		},
		Entry("images-repo triggers granular mode",
			"registry.example.com/images", "", nil, nil, true),
		Entry("meta-repo triggers granular mode",
			"", "registry.example.com/meta", nil, nil, true),
		Entry("no granular flags means legacy mode",
			"", "", nil, nil, false),
	)

	It("should be in legacy mode when only --repo is set", func() {
		cmdData := &CmdData{
			Repo:       NewRepoData("repo", RepoDataOptions{}),
			ImagesRepo: &[]string{},
			MetaRepo:   NewRepoData("meta-repo", RepoDataOptions{OnlyAddress: true}),
			CacheFrom:  &[]string{},
			CacheTo:    &[]string{},
		}
		repoAddr := "registry.example.com/repo"
		cmdData.Repo.Address = &repoAddr
		emptyAddr := ""
		cmdData.MetaRepo.Address = &emptyAddr

		Expect(getGranularMode(cmdData)).To(BeFalse())
	})
})
