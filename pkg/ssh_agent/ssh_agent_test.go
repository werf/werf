package ssh_agent

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/werf/werf/v2/pkg/werf"
)

const (
	longPath = "/tmp/werf-test-agent-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestLinuxFallback_with_SSHenv(t *testing.T) {
	if os.Getenv(SSHAuthSockEnv) == "" {
		t.Skip("Skipping test because SSH_AUTH_SOCK is not set")
	}
	ctx := context.Background()
	err := testInit(ctx, t, false, false)
	assert.NoError(t, err)

	valid, err := validatesystemAgentSock(SSHAuthSock)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestLinuxFallback_without_SSHenv_wildTmpPath(t *testing.T) {
	ctx := context.Background()
	err := testInit(ctx, t, false, false)
	assert.NoError(t, err)

	valid, err := validatesystemAgentSock(SSHAuthSock)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestLinuxFallback_with_SSHenv_wildTmpPath(t *testing.T) {
	if os.Getenv(SSHAuthSockEnv) == "" {
		t.Skip("Skipping test because SSH_AUTH_SOCK is not set")
	}
	ctx := context.Background()
	err := testInit(ctx, t, false, false)
	assert.NoError(t, err)

	valid, err := validatesystemAgentSock(SSHAuthSock)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestLinuxFallback_withLongSSHenv(t *testing.T) {
	ctx := context.Background()
	os.Setenv(SSHAuthSockEnv, longPath)
	err := testInit(ctx, t, false, false)
	assert.Error(t, err)
}

func testInit(ctx context.Context, t *testing.T, unsetSSHenv, wildTmpPath bool) error {
	if unsetSSHenv {
		os.Unsetenv("SSH_AUTH_SOCK")
	}
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("Skipping test on non-linux OS")
	}

	home, _ := os.UserHomeDir()

	testPath := "/tmp/werf-test-agent"
	if wildTmpPath {
		testPath = longPath
	}

	err := os.MkdirAll(testPath, os.ModePerm)
	if err != nil {
		return err
	}
	defer os.RemoveAll(testPath)

	err = werf.Init(testPath, home)
	if err != nil {
		return err
	}

	err = Init(ctx, []string{})
	if err != nil {
		return err
	}

	return nil
}
