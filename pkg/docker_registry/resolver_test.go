package docker_registry_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/docker_registry"
)

type entry struct {
	imagesRepoAddress string
	expectation       string
}

var _ = DescribeTable("resolve implementation name", func(entry entry) {
	resolvedImplementation, err := docker_registry.ResolveImplementation(entry.imagesRepoAddress, "")
	Ω(err).ShouldNot(HaveOccurred())

	Ω(resolvedImplementation).Should(Equal(entry.expectation))
},
	Entry("ecr", entry{
		imagesRepoAddress: "123456789012.dkr.ecr.test.amazonaws.com/repo",
		expectation:       "ecr",
	}),
	Entry("acr", entry{
		imagesRepoAddress: "test.azurecr.io/repo",
		expectation:       "acr",
	}),
	Entry("default", entry{
		imagesRepoAddress: "localhost:5000/repo",
		expectation:       "default",
	}),
	Entry("dockerhub", entry{
		imagesRepoAddress: "account/repo",
		expectation:       "dockerhub",
	}),
	Entry("gcr", entry{
		imagesRepoAddress: "container.cloud.google.com/registry/repo",
		expectation:       "gcr",
	}),
	Entry("github", entry{
		imagesRepoAddress: "docker.pkg.github.com/account/project/package",
		expectation:       "github",
	}),
	Entry("harbor", entry{
		imagesRepoAddress: "harbor.company.com/project/repo",
		expectation:       "harbor",
	}),
	Entry("quay", entry{
		imagesRepoAddress: "quay.io/account/repo",
		expectation:       "quay",
	}),
)
