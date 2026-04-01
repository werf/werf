//go:build ai_tests

package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	dockerbuildkit "github.com/docker/docker/client/buildkit"
	buildkitclient "github.com/moby/buildkit/client"
	buildkitexptypes "github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/stretchr/testify/require"
)

func TestAI_BuildKitConnection(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" && !fileExists("/var/run/docker.sock") {
		t.Skip("Docker daemon not available")
	}

	ctx := context.Background()

	if err := Init(ctx, InitOptions{}); err != nil {
		t.Skipf("Failed to initialize Docker: %v", err)
	}

	ctx, err := NewContext(ctx)
	require.NoError(t, err)

	// Connect to Docker's embedded buildkitd via Docker API hijack (/grpc + /session).
	bk, err := buildkitclient.New(ctx, "", dockerbuildkit.ClientOpts(apiCli(ctx))...)
	require.NoError(t, err)
	defer bk.Close()

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	require.NoError(t, bk.Wait(waitCtx))

	buildDir := t.TempDir()
	dockerfilePath := filepath.Join(buildDir, "Dockerfile")
	require.NoError(t, os.WriteFile(dockerfilePath, []byte("FROM scratch\n"), 0o644))

	solveCtx, solveCancel := context.WithTimeout(ctx, 30*time.Second)
	defer solveCancel()

	response, err := bk.Solve(solveCtx, nil, buildkitclient.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": "Dockerfile",
		},
		LocalDirs: map[string]string{
			"context":    buildDir,
			"dockerfile": buildDir,
		},
		Exports: []buildkitclient.ExportEntry{
			{
				Type: buildkitclient.ExporterImage,
				Attrs: map[string]string{
					string(buildkitexptypes.OptKeyName):  "werf-buildkit-spike:latest",
					string(buildkitexptypes.OptKeyStore): "true",
					string(buildkitexptypes.OptKeyPush):  "false",
				},
			},
		},
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.ExporterResponse["containerimage.digest"])
}
