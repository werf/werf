package buildkit

import (
	"fmt"

	"github.com/docker/cli/cli/config"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
)

type SessionAttachablesOptions struct {
	DockerConfigDir string
	SSHAgentSocks   []sshprovider.AgentConfig
	Secrets         []secretsprovider.Source
}

func SessionAttachables(opts SessionAttachablesOptions) ([]session.Attachable, error) {
	dockerConfig, err := config.Load(opts.DockerConfigDir)
	if err != nil {
		return nil, fmt.Errorf("load docker config from %q: %w", opts.DockerConfigDir, err)
	}

	attachables := []session.Attachable{
		authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
			AuthConfigProvider: authprovider.LoadAuthConfig(dockerConfig),
		}),
	}

	if len(opts.SSHAgentSocks) > 0 {
		sshProvider, err := sshprovider.NewSSHAgentProvider(opts.SSHAgentSocks)
		if err != nil {
			return nil, fmt.Errorf("create ssh agent provider: %w", err)
		}
		attachables = append(attachables, sshProvider)
	}

	if len(opts.Secrets) > 0 {
		store, err := secretsprovider.NewStore(opts.Secrets)
		if err != nil {
			return nil, fmt.Errorf("create secrets store: %w", err)
		}
		attachables = append(attachables, secretsprovider.NewSecretProvider(store))
	}

	return attachables, nil
}
