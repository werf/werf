package docker

import (
	"fmt"
	"net"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

func LocalDockerRegistryRun() (string, string, string) {
	containerName := fmt.Sprintf("werf_test_docker_registry-%s", utils.GetRandomString(10))
	imageName := "registry:2"

	if exist := IsImageExist(imageName); !exist {
		err := Pull(imageName)
		Ω(err).ShouldNot(HaveOccurred(), "docker pull "+imageName)
	}

	port, err := getFreeEphemeralPort()
	Ω(err).ShouldNot(HaveOccurred())

	dockerCliRunArgs := []string{
		"-d",
		"-p", fmt.Sprintf("%d:5000", port),
		"-e", "REGISTRY_STORAGE_DELETE_ENABLED=true",
		"--name", containerName,
		imageName,
	}
	err = CliRun(dockerCliRunArgs...)
	Ω(err).ShouldNot(HaveOccurred(), "docker run "+strings.Join(dockerCliRunArgs, " "))

	inspect := ContainerInspect(containerName)
	Ω(inspect.NetworkSettings).ShouldNot(BeNil())
	Ω(inspect.NetworkSettings.IPAddress).ShouldNot(BeEmpty())
	registryInternalAddress := fmt.Sprintf("%s:%d", inspect.NetworkSettings.IPAddress, 5000)

	registryLocalAddress := fmt.Sprintf("localhost:%d", port)
	registryWithScheme := fmt.Sprintf("http://%s", registryLocalAddress)

	utils.WaitTillHostReadyToRespond(registryWithScheme, utils.DefaultWaitTillHostReadyToRespondMaxAttempts)

	return registryLocalAddress, registryInternalAddress, containerName
}

func getFreeEphemeralPort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	return port, nil
}
