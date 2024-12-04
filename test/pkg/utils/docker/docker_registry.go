package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/util"
	"github.com/werf/werf/v2/test/pkg/utils"
)

func LocalDockerRegistryRun() (string, string, string) {
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
	err = CliRun(dockerCliRunArgs...)
	Expect(err).ShouldNot(HaveOccurred(), "docker run "+strings.Join(dockerCliRunArgs, " "))

	inspect := ContainerInspect(containerName)
	Expect(inspect.NetworkSettings).ShouldNot(BeNil())
	Expect(inspect.NetworkSettings.IPAddress).ShouldNot(BeEmpty())
	registryInternalAddress := fmt.Sprintf("%s:%d", inspect.NetworkSettings.IPAddress, 5000)

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

// CatalogResponse структура для ответа от Docker Registry API
type CatalogResponse struct {
	Repositories []string `json:"repositories"`
}

func fetchCatalogWithContext(ctx context.Context, registryURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://%s/v2/_catalog", registryURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var catalog CatalogResponse
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return catalog.Repositories, nil
}

func LookUpForRepositories(ctx context.Context, registryURL string, repositories []string) bool {
	repos, err := fetchCatalogWithContext(ctx, registryURL)
	Expect(err).ShouldNot(HaveOccurred())

	reposMap := util.SliceToMapWithValue(repos, struct{}{})

	for _, repo := range repositories {
		if _, ok := reposMap[repo]; !ok {
			return false
		}
	}

	return true
}

type TagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func fetchTagsWithContext(ctx context.Context, registryURL, repository string) ([]string, error) {
	url := fmt.Sprintf("http://%s/v2/%s/tags/list", registryURL, repository)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tagsResponse TagsResponse
	if err := json.Unmarshal(body, &tagsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return tagsResponse.Tags, nil
}

func LookUpForTag(ctx context.Context, registryURL, repository, tag string) bool {
	tags, err := fetchTagsWithContext(ctx, registryURL, repository)
	Expect(err).ShouldNot(HaveOccurred())

	for _, t := range tags {
		if t == tag {
			return true
		}
	}

	return false
}
