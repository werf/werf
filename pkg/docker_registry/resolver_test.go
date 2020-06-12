package docker_registry_test

import (
	"github.com/werf/werf/pkg/docker_registry"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type entry struct {
	imagesRepoAddress          string
	expectedImplementationName string
	expectedRepoMode           string
}

var _ = DescribeTable("resolve implementation name and repo mode", func(entry entry) {
	resolvedImplementation, err := docker_registry.ResolveImplementation(entry.imagesRepoAddress, "")
	Ω(err).ShouldNot(HaveOccurred())

	Ω(resolvedImplementation).Should(Equal(entry.expectedImplementationName))

	dockerRegistry, err := docker_registry.NewDockerRegistry(entry.imagesRepoAddress, resolvedImplementation, docker_registry.DockerRegistryOptions{})
	Ω(err).ShouldNot(HaveOccurred())

	resolvedRepoMode, err := dockerRegistry.ResolveRepoMode(entry.imagesRepoAddress, "")
	Ω(err).ShouldNot(HaveOccurred())

	Ω(resolvedRepoMode).Should(Equal(entry.expectedRepoMode))
},
	Entry("[ecr] registry -> multirepo", entry{
		imagesRepoAddress:          "123456789012.dkr.ecr.test.amazonaws.com",
		expectedImplementationName: "ecr",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[ecr] repository -> multirepo", entry{
		imagesRepoAddress:          "123456789012.dkr.ecr.test.amazonaws.com/test",
		expectedImplementationName: "ecr",
		expectedRepoMode:           "multirepo",
	}),

	Entry("[acr] registry -> multirepo", entry{
		imagesRepoAddress:          "test.azurecr.io",
		expectedImplementationName: "acr",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[acr] repository -> multirepo", entry{
		imagesRepoAddress:          "test.azurecr.io/test",
		expectedImplementationName: "acr",
		expectedRepoMode:           "multirepo",
	}),

	Entry("[default] registry -> multirepo", entry{
		imagesRepoAddress:          "localhost:5000",
		expectedImplementationName: "default",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[default] repository -> multirepo", entry{
		imagesRepoAddress:          "localhost:5000/test",
		expectedImplementationName: "default",
		expectedRepoMode:           "multirepo",
	}),

	Entry("[dockerhub] account -> multirepo", entry{
		imagesRepoAddress:          "account",
		expectedImplementationName: "dockerhub",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[dockerhub] repository -> monorepo", entry{
		imagesRepoAddress:          "account/repo",
		expectedImplementationName: "dockerhub",
		expectedRepoMode:           "monorepo",
	}),

	Entry("[gcr] registry -> multirepo", entry{
		imagesRepoAddress:          "container.cloud.google.com/registry",
		expectedImplementationName: "gcr",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[gcr] repository -> multirepo", entry{
		imagesRepoAddress:          "container.cloud.google.com/registry/repo",
		expectedImplementationName: "gcr",
		expectedRepoMode:           "multirepo",
	}),

	Entry("[github] registry -> multirepo", entry{
		imagesRepoAddress:          "docker.pkg.github.com/account/project",
		expectedImplementationName: "github",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[github] repository -> monorepo", entry{
		imagesRepoAddress:          "docker.pkg.github.com/account/project/package",
		expectedImplementationName: "github",
		expectedRepoMode:           "monorepo",
	}),

	Entry("[harbor] registry -> multirepo", entry{
		imagesRepoAddress:          "harbor.company.com/project",
		expectedImplementationName: "harbor",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[harbor] repository -> multirepo", entry{
		imagesRepoAddress:          "harbor.company.com/project/repo",
		expectedImplementationName: "harbor",
		expectedRepoMode:           "multirepo",
	}),

	Entry("[quay] registry -> multirepo", entry{
		imagesRepoAddress:          "quay.io/account",
		expectedImplementationName: "quay",
		expectedRepoMode:           "multirepo",
	}),
	Entry("[quay] repository -> monorepo", entry{
		imagesRepoAddress:          "quay.io/account/repo",
		expectedImplementationName: "quay",
		expectedRepoMode:           "monorepo",
	}),
)
