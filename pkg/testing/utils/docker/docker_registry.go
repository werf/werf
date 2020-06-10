package docker

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/testing/utils"
)

func LocalDockerRegistryRun() (string, string) {
	containerName := fmt.Sprintf("werf_test_docker_registry-%s", utils.GetRandomString(10))
	imageName := "flant/werf-test:registry"

	if exist := IsImageExist(imageName); !exist {
		err := Pull(imageName)
		Ω(err).ShouldNot(HaveOccurred(), "docker pull "+imageName)
	}

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
