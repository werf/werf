package docker

import (
	"bytes"
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/stretchr/testify/require"
)

func TestAI_RequiredDockerCliCommandsExist(t *testing.T) {
	cli, err := command.NewDockerCli(
		command.WithOutputStream(&bytes.Buffer{}),
		command.WithErrorStream(&bytes.Buffer{}),
	)
	require.NoError(t, err, "failed to create docker CLI")

	err = validateRequiredCliCommands(cli)
	require.NoError(t, err, "required docker CLI commands are missing")
}
