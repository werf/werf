package buildkit

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestAI_SessionAttachables_AuthOnly(t *testing.T) {
	attachables, err := SessionAttachables(SessionAttachablesOptions{DockerConfigDir: t.TempDir()})
	require.NoError(t, err)
	assert.Len(t, attachables, 1)
}

func TestAI_SessionAttachables_WithSSHAndSecrets(t *testing.T) {
	tmpDir := t.TempDir()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	pemBlock, err := ssh.MarshalPrivateKey(privateKey, "")
	require.NoError(t, err)
	keyFile := filepath.Join(tmpDir, "id_ed25519")
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(pemBlock), 0o600))

	secretFile := filepath.Join(tmpDir, "secret")
	require.NoError(t, os.WriteFile(secretFile, []byte("value"), 0o600))

	attachables, err := SessionAttachables(SessionAttachablesOptions{
		DockerConfigDir: tmpDir,
		SSHAgentSocks:   []sshprovider.AgentConfig{{Paths: []string{keyFile}}},
		Secrets:         []secretsprovider.Source{{ID: "mysecret", FilePath: secretFile}},
	})
	require.NoError(t, err)
	assert.Len(t, attachables, 3)
}

func TestAI_SessionAttachables_SecretWithoutIDFails(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(secretFile, []byte("value"), 0o600))

	_, err := SessionAttachables(SessionAttachablesOptions{
		DockerConfigDir: t.TempDir(),
		Secrets:         []secretsprovider.Source{{FilePath: secretFile}},
	})
	require.Error(t, err)
}
