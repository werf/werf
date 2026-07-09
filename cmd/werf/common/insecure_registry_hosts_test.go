package common

import (
	"context"
	"testing"
)

func TestGetInsecureRegistryHosts_SkipWhenInsecureRegistryEnabled(t *testing.T) {
	cmdData := &CmdData{
		InsecureRegistry:      boolPtr(true),
		SkipTlsVerifyRegistry: boolPtr(false),
	}

	hosts, err := GetInsecureRegistryHosts(context.Background(), cmdData)
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

	hosts, err := GetInsecureRegistryHosts(context.Background(), cmdData)
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

	_, err := CreateDockerRegistryWithInsecureHosts(context.Background(), cmdData, "registry.example.com/project")
	if err != nil {
		t.Fatalf("expected no error creating registry client with global insecure mode, got: %v", err)
	}
}

func TestCreateDockerRegistryWithInsecureHosts_SkipWhenSkipTLSVerifyEnabled(t *testing.T) {
	cmdData := &CmdData{
		InsecureRegistry:      boolPtr(false),
		SkipTlsVerifyRegistry: boolPtr(true),
	}

	_, err := CreateDockerRegistryWithInsecureHosts(context.Background(), cmdData, "registry.example.com/project")
	if err != nil {
		t.Fatalf("expected no error creating registry client with skip tls verify mode, got: %v", err)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
