package common

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/container_backend"
)

var _ = ginkgo.Describe("Multiple final repo addresses", func() {
	ginkgo.BeforeEach(clearListEnv)

	ginkgo.DescribeTable("parses repeated flags and environment defaults",
		func(env map[string]string, args, expected []string) {
			for key, value := range env {
				ginkgo.GinkgoT().Setenv(key, value)
			}
			command := &cobra.Command{}
			repo := NewRepoData("final-repo", RepoDataOptions{MultipleAddresses: true})
			repo.SetupCmd(command)
			gomega.Expect(command.ParseFlags(args)).To(gomega.Succeed())
			gomega.Expect(repo.Addresses).To(gomega.Equal(expected))
		},
		ginkgo.Entry("no destination", nil, nil, []string(nil)),
		ginkgo.Entry("legacy environment", map[string]string{"WERF_FINAL_REPO": "example.com/one"}, nil, []string{"example.com/one"}),
		ginkgo.Entry("ordered environment", map[string]string{
			"WERF_FINAL_REPO": "example.com/first", "WERF_FINAL_REPO_2": "example.com/third", "WERF_FINAL_REPO_1": "example.com/second",
		}, nil, []string{"example.com/first", "example.com/second", "example.com/third"}),
		ginkgo.Entry("repeated flags", nil, []string{"--final-repo=example.com/one", "--final-repo=example.com/two"}, []string{"example.com/one", "example.com/two"}),
		ginkgo.Entry("flags replace environment", map[string]string{"WERF_FINAL_REPO": "example.com/env", "WERF_FINAL_REPO_1": "example.com/env-one"}, []string{"--final-repo=example.com/cli"}, []string{"example.com/cli"}),
		ginkgo.Entry("credentials are not repository addresses", map[string]string{
			"WERF_FINAL_REPO": "example.com/one", "WERF_FINAL_REPO_CONTAINER_REGISTRY": "default",
			"WERF_FINAL_REPO_DOCKER_HUB_USERNAME": "user", "WERF_FINAL_REPO_DOCKER_HUB_PASSWORD": "password",
			"WERF_FINAL_REPO_DOCKER_HUB_TOKEN": "token", "WERF_FINAL_REPO_GITHUB_TOKEN": "token",
			"WERF_FINAL_REPO_HARBOR_USERNAME": "user", "WERF_FINAL_REPO_HARBOR_PASSWORD": "password", "WERF_FINAL_REPO_QUAY_TOKEN": "token",
		}, nil, []string{"example.com/one"}),
		ginkgo.Entry("legacy registry options are not repository addresses", map[string]string{
			"WERF_FINAL_REPO": "example.com/one", "WERF_FINAL_REPO_IMPLEMENTATION": "default",
			"WERF_FINAL_REPO_SELECTEL_USERNAME": "user", "WERF_FINAL_REPO_SELECTEL_PASSWORD": "review-fixture-secret",
		}, nil, []string{"example.com/one"}),
		ginkgo.Entry("only numeric suffixes select repositories", map[string]string{
			"WERF_FINAL_REPO": "example.com/one", "WERF_FINAL_REPO_1": "example.com/two",
			"WERF_FINAL_REPO_": "ignored", "WERF_FINAL_REPO_1_PASSWORD": "ignored",
			"WERF_FINAL_REPO_BACKUP": "ignored", "WERF_FINAL_REPO_2A": "ignored",
		}, nil, []string{"example.com/one", "example.com/two"}),
	)

	ginkgo.DescribeTable("constructs final storages from command options",
		func(ctx ginkgo.SpecContext, multiple bool) {
			address, _ := startRecordingRegistry("project")
			data := storageManagerCmdData("--repo", address, "--insecure-registry")
			command := &cobra.Command{}
			args := []string{"--final-repo=" + address}
			expected := []string{address}
			if multiple {
				data.FinalRepo = NewRepoData("final-repo", RepoDataOptions{MultipleAddresses: true})
				data.FinalRepo.SetupCmd(command)
				args = append(args, "--final-repo="+address, "--final-repo="+address+"/other", "--final-repo="+address)
				expected = append(expected, address+"/other")
			} else {
				SetupFinalRepo(data, command)
			}
			gomega.Expect(command.ParseFlags(args)).To(gomega.Succeed())
			storages, err := GetFinalStagesStorageList(ctx, &container_backend.DockerServerBackend{}, data)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(storages).To(gomega.HaveLen(len(expected)))
			for i, finalStorage := range storages {
				gomega.Expect(finalStorage.Address()).To(gomega.Equal(expected[i]))
			}
		},
		ginkgo.Entry("keeps the destination of single-address commands", false),
		ginkgo.Entry("deduplicates multiple destinations in order", true),
	)
})
