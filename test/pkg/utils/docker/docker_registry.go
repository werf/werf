package docker

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

func LocalDockerRegistryRun() (string, string, string) {
	containerName := fmt.Sprintf("werf_test_docker_registry-%s", utils.GetRandomString(10))
	imageName := "registry:2"

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

	inspect := ContainerInspect(containerName)
	Ω(inspect.NetworkSettings).ShouldNot(BeNil())
	Ω(inspect.NetworkSettings.IPAddress).ShouldNot(BeEmpty())
	registryInternalAddress := fmt.Sprintf("%s:%d", inspect.NetworkSettings.IPAddress, 5000)

	registryLocalAddress := fmt.Sprintf("localhost:%s", ContainerHostPort(containerName, "5000/tcp"))
	registryWithScheme := fmt.Sprintf("http://%s", registryLocalAddress)

	utils.WaitTillHostReadyToRespond(registryWithScheme, utils.DefaultWaitTillHostReadyToRespondMaxAttempts)

	return registryLocalAddress, registryInternalAddress, containerName
}
