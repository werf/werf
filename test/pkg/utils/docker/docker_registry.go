package docker

import (
	"context"
	"fmt"
	"net"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

func LocalDockerRegistryRun(ctx context.Context) (string, string, string) {
	containerName := fmt.Sprintf("werf_test_docker_registry-%s", utils.GetRandomString(10))
	imageName := "registry:2"

	port, err := getFreeEphemeralPort()
	Expect(err).ShouldNot(HaveOccurred())

	dockerCliRunArgs := []string{
		"-d",
		"-p", fmt.Sprintf("%d:5000", port),
		"-e", "REGISTRY_STORAGE_DELETE_ENABLED=true",
		"--name", containerName,
		imageName,
	}
	err = CliRun(ctx, dockerCliRunArgs...)
	Expect(err).ShouldNot(HaveOccurred(), "docker run "+strings.Join(dockerCliRunArgs, " "))

	inspect := ContainerInspect(ctx, containerName)
	Expect(inspect.NetworkSettings).ShouldNot(BeNil())
	Expect(inspect.NetworkSettings.Networks["bridge"].IPAddress).ShouldNot(BeEmpty())
	registryInternalAddress := fmt.Sprintf("%s:%d", inspect.NetworkSettings.Networks["bridge"].IPAddress, 5000)

	registryLocalAddress := fmt.Sprintf("localhost:%d", port)
	registryWithScheme := fmt.Sprintf("http://%s", registryLocalAddress)

	utils.WaitTillHostReadyToRespond(registryWithScheme, utils.DefaultWaitTillHostReadyToRespondMaxAttempts)

	return registryLocalAddress, registryInternalAddress, containerName
}

// Listen on a random port using TCP protocol on localhost and return the free ephemeral port number.
func getFreeEphemeralPort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	return port, nil
}
