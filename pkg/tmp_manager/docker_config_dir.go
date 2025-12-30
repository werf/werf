package tmp_manager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/otiai10/copy"
)

func CreateDockerConfigDir(ctx context.Context, fromDockerConfig string) (string, error) {
	newDir, err := newTmpDir(dockerConfigDirPrefix)
	if err != nil {
		return "", err
	}

	if err = registrator.queueRegistration(ctx, newDir, filepath.Join(getCreatedTmpDirs(), dockerConfigsServiceDir)); err != nil {
		return "", fmt.Errorf("unable to queue GC registration: %w", err)
	}

	if err = os.Chmod(newDir, 0o700); err != nil {
		return "", err
	}

	if _, err = os.Stat(fromDockerConfig); errors.Is(err, os.ErrNotExist) {
		return newDir, nil // Nothing to copy
	} else if err != nil {
		return "", err
	}

	// Some Docker's configurations:
	// - `~/.docker/run` — Runtime directory holding temporary files (e.g., Unix sockets, PID/lock/state files) for user-scoped Docker components (often Docker Desktop/proxies). Not used for images; typically recreated on restart.
	// - `~/.docker/mutagen` — Mutagen data (file sync / accelerated sharing): sync session state and metadata, caches. Removing it resets/loses Mutagen sync sessions/state.
	// - `~/.docker/desktop` — Docker Desktop user data and settings (configs, integrations, internal service files; exact contents vary by version). Deleting it usually resets Desktop settings.
	// - `~/.docker/contexts` — Docker Contexts: connection configurations to Docker Engine (local/remote), endpoints, and TLS certificates/keys for those connections. Deleting it removes contexts and may break access to remote hosts.
	// - `~/.docker/trust` — Docker Content Trust / Notary data: image signing keys and trust metadata. Losing it can prevent signing/updating trusted repositories.
	// - `~/.docker/cli-plugins` — User-installed Docker CLI plugins (executables named `docker-<plugin>`, e.g., `docker-compose`). Removing a plugin removes the corresponding subcommand.
	// - `~/.docker/buildx` — `docker buildx` data: builder instance configuration and metadata, BuildKit-related settings, local state. Deleting it typically requires recreating builders.
	// - `~/.docker/config.json` — Main Docker CLI config: client settings, proxies, parameters, and registry authentication (`auths`) or credential store/helper configuration (`credsStore`/`credHelpers`). Deleting it logs you out and resets client settings.
	// - `~/.docker/features.json` — Feature flags/toggles for Docker and/or plugins (uncommon; depends on version/distribution). Typically controls enabling/disabling specific client/tool capabilities.
	// - `~/.docker/scout` – location for the docker-scout CLI plugin binary when installed manually; referenced via cliPluginsExtraDirs in config.json.

	// Define options to skip specific directories
	dockerPathsToSkip := []string{"cli-plugins", "buildx", "machine", "desktop", "run", "mutagen", "scout"}

	options := copy.Options{
		Skip: func(srcInfo os.FileInfo, src, dst string) (bool, error) {
			return slices.Contains(dockerPathsToSkip, srcInfo.Name()), nil
		},
	}

	if err = copy.Copy(fromDockerConfig, newDir, options); err != nil {
		return "", fmt.Errorf("unable to copy %q to %q: %w", fromDockerConfig, newDir, err)
	}

	return newDir, nil
}
