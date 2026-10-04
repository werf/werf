package common

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/werf/werf/v2/pkg/buildah"
)

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
