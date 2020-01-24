package docker

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/flant/go-containerregistry/pkg/authn"
	"github.com/flant/go-containerregistry/pkg/name"
	"github.com/flant/go-containerregistry/pkg/v1/remote"

	"github.com/flant/werf/pkg/testing/utils"
)

func RegistryRepositoryList(reference string) []string {
	repo, err := name.NewRepository(reference, name.WeakValidation)
	Ω(err).ShouldNot(HaveOccurred(), fmt.Sprintf("parsing repo %q: %v", reference, err))

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil && strings.Contains(err.Error(), "NAME_UNKNOWN") {
		return []string{}
	}

	Ω(err).ShouldNot(HaveOccurred(), fmt.Sprintf("reading tags for %q: %v", repo, err))
	return tags
}

func LocalDockerRegistryRun() (string, string) {
	containerName := fmt.Sprintf("werf_test_docker_registry-%s", utils.GetRandomString(10))
	imageName := "flant/werf-test:registry"

	dockerCliRunArgs := []string{
		"-d",
		"-p", ":5000",
		"-e", "REGISTRY_STORAGE_DELETE_ENABLED=true",
		"--name", containerName,
		imageName,
	}
	err := CliRun(dockerCliRunArgs...)
	Ω(err).ShouldNot(HaveOccurred(), "docker run "+strings.Join(dockerCliRunArgs, " "))

	registry := fmt.Sprintf("localhost:%s", ContainerHostPort(containerName, "5000/tcp"))
	registryWithScheme := fmt.Sprintf("http://%s", registry)

	utils.WaitTillHostReadyToRespond(registryWithScheme, utils.DefaultWaitTillHostReadyToRespondMaxAttempts)

	return registry, containerName
}
