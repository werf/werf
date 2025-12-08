package docker

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

func LocalDockerRegistryRun(ctx context.Context) (string, string, string) {
	containerName := fmt.Sprintf("werf_test_docker_registry-%s", utils.GetRandomString(10))
	imageName := "registry:2"

	dockerCliRunArgs := []string{
		"-d",
		"-p", "0:5000",
		"-e", "REGISTRY_STORAGE_DELETE_ENABLED=true",
		"--name", containerName,
		imageName,
	}

	err := CliRun(ctx, dockerCliRunArgs...)
	Expect(err).ShouldNot(HaveOccurred(), "docker run "+strings.Join(dockerCliRunArgs, " "))

	inspect := ContainerInspect(ctx, containerName)
	Expect(inspect.NetworkSettings).ShouldNot(BeNil())
	Expect(inspect.NetworkSettings.Networks["bridge"].IPAddress).ShouldNot(BeEmpty())
	registryInternalAddress := fmt.Sprintf("%s:%d", inspect.NetworkSettings.Networks["bridge"].IPAddress, 5000)

	portBindings := inspect.NetworkSettings.Ports["5000/tcp"]
	Expect(portBindings).ShouldNot(BeNil())
	Expect(portBindings).ShouldNot(BeEmpty())

	hostPort := portBindings[0].HostPort
	registryLocalAddress := fmt.Sprintf("localhost:%s", hostPort)
	registryWithScheme := fmt.Sprintf("http://%s", registryLocalAddress)

	utils.WaitTillHostReadyToRespond(registryWithScheme, utils.DefaultWaitTillHostReadyToRespondMaxAttempts)

	return registryLocalAddress, registryInternalAddress, containerName
}
