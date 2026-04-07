package common

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/werf/werf/v2/pkg/buildah"
)

func TestGetInsecureRegistryHosts_SkipWhenInsecureRegistryEnabled(t *testing.T) {
	cmdData := &CmdData{
		InsecureRegistry:      boolPtr(true),
		SkipTlsVerifyRegistry: boolPtr(false),
	}

	hosts, err := GetInsecureRegistryHosts(context.Background(), cmdData, buildah.ModeDisabled)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(hosts) != 0 {
		t.Fatalf("expected empty insecure hosts list, got: %v", hosts)
	}
}

func TestGetInsecureRegistryHosts_SkipWhenSkipTLSVerifyEnabled(t *testing.T) {
	cmdData := &CmdData{
		InsecureRegistry:      boolPtr(false),
		SkipTlsVerifyRegistry: boolPtr(true),
	}

	hosts, err := GetInsecureRegistryHosts(context.Background(), cmdData, buildah.ModeNative)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(hosts) != 0 {
		t.Fatalf("expected empty insecure hosts list, got: %v", hosts)
	}
}

func TestCreateDockerRegistryWithInsecureHosts_SkipWhenInsecureRegistryEnabled(t *testing.T) {
	cmdData := &CmdData{
		InsecureRegistry:      boolPtr(true),
		SkipTlsVerifyRegistry: boolPtr(false),
	}

	_, err := CreateDockerRegistryWithInsecureHosts(context.Background(), cmdData, "registry.example.com/project", buildah.ModeNative)
	if err != nil {
		t.Fatalf("expected no error creating registry client with global insecure mode, got: %v", err)
	}
}

func TestCreateDockerRegistryWithInsecureHosts_SkipWhenSkipTLSVerifyEnabled(t *testing.T) {
	cmdData := &CmdData{
		InsecureRegistry:      boolPtr(false),
		SkipTlsVerifyRegistry: boolPtr(true),
	}

	_, err := CreateDockerRegistryWithInsecureHosts(context.Background(), cmdData, "registry.example.com/project", buildah.ModeNative)
	if err != nil {
		t.Fatalf("expected no error creating registry client with skip tls verify mode, got: %v", err)
	}
}

func TestGetContainerRegistryMirror_MergesWerfEnvMirrorsBeforeBuildahConfigMirrors(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldMirror := os.Getenv("WERF_CONTAINER_REGISTRY_MIRROR_1")
	oldRegistriesConf := os.Getenv("CONTAINERS_REGISTRIES_CONF")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-registries.conf")
	if err := os.WriteFile(configPath, []byte(`
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "config-mirror.example.com"
`), 0o644); err != nil {
		t.Fatalf("write registries.conf: %v", err)
	}

	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("set HOME: %v", err)
	}
	if err := os.Setenv("CONTAINERS_REGISTRIES_CONF", configPath); err != nil {
		t.Fatalf("set CONTAINERS_REGISTRIES_CONF: %v", err)
	}
	if err := os.Setenv("WERF_CONTAINER_REGISTRY_MIRROR_1", "env-mirror.example.com"); err != nil {
		t.Fatalf("set WERF_CONTAINER_REGISTRY_MIRROR_1: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", oldHome)
		if oldMirror == "" {
			_ = os.Unsetenv("WERF_CONTAINER_REGISTRY_MIRROR_1")
		} else {
			_ = os.Setenv("WERF_CONTAINER_REGISTRY_MIRROR_1", oldMirror)
		}
		if oldRegistriesConf == "" {
			_ = os.Unsetenv("CONTAINERS_REGISTRIES_CONF")
		} else {
			_ = os.Setenv("CONTAINERS_REGISTRIES_CONF", oldRegistriesConf)
		}
	}()

	cmdData := &CmdData{
		ContainerRegistryMirror: &[]string{},
	}
	mirrors, err := GetContainerRegistryMirror(context.Background(), cmdData, buildah.ModeNative)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(mirrors) != 2 {
		t.Fatalf("expected 2 mirrors, got: %v", mirrors)
	}
	if mirrors[0] != "https://env-mirror.example.com" {
		t.Fatalf("expected env mirror first, got: %v", mirrors)
	}
	if mirrors[1] != "https://config-mirror.example.com" {
		t.Fatalf("expected config mirror second, got: %v", mirrors)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
