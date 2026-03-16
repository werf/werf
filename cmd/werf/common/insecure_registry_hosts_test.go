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

func boolPtr(v bool) *bool {
	return &v
}
