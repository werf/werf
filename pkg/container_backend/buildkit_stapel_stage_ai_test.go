package container_backend

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonistiigi/fsutil"
)

var testPlatform = ocispecs.Platform{OS: "linux", Architecture: "amd64"}

func marshalStapelOps(t *testing.T, state llb.State) []*pb.Op {
	t.Helper()
	def, err := state.Marshal(context.Background())
	require.NoError(t, err)

	var ops []*pb.Op
	for _, raw := range def.Def {
		op := &pb.Op{}
		require.NoError(t, op.UnmarshalVT(raw))
		ops = append(ops, op)
	}
	return ops
}

func stapelFileActions(t *testing.T, ops []*pb.Op) []*pb.FileAction {
	t.Helper()
	var actions []*pb.FileAction
	for _, op := range ops {
		if file := op.GetFile(); file != nil {
			actions = append(actions, file.Actions...)
		}
	}
	return actions
}

type stubRefPinner struct{}

func (stubRefPinner) ResolvePinnedRef(_ context.Context, ref string, _ *ocispecs.Platform) (string, []byte, error) {
	return ref + "@sha256:2222222222222222222222222222222222222222222222222222222222222222", nil, nil
}

func TestAI_BuildkitStapel_DependencyImports(t *testing.T) {
	state, err := applyStapelDependencyImports(context.Background(), llb.Scratch(), stubRefPinner{}, []DependencyImportSpec{
		{
			ImageName:    "registry.example.com/project:dep",
			FromPath:     "/from",
			ToPath:       "/to",
			IncludePaths: []string{"**/*.txt"},
			ExcludePaths: []string{"skip"},
			Owner:        "1000",
			Group:        "app",
		},
	}, testPlatform)
	require.NoError(t, err)

	ops := marshalStapelOps(t, state)

	var imageFound bool
	for _, op := range ops {
		if src := op.GetSource(); src != nil && src.Identifier == "docker-image://registry.example.com/project:dep@sha256:2222222222222222222222222222222222222222222222222222222222222222" {
			imageFound = true
		}
	}
	assert.True(t, imageFound)

	actions := stapelFileActions(t, ops)
	require.Len(t, actions, 1)
	copyAction := actions[0].GetCopy()
	require.NotNil(t, copyAction)
	assert.Equal(t, "/from", copyAction.Src)
	assert.Equal(t, "/to", copyAction.Dest)
	assert.Equal(t, []string{"**/*.txt"}, copyAction.IncludePatterns)
	assert.Equal(t, []string{"skip"}, copyAction.ExcludePatterns)
	assert.True(t, copyAction.DirCopyContents)
	require.NotNil(t, copyAction.Owner)
	assert.EqualValues(t, 1000, copyAction.Owner.User.GetByID())
	assert.Equal(t, "app", copyAction.Owner.Group.GetByName().Name)
}

func makeTestTar(t *testing.T, files map[string]string) io.ReadCloser {
	t.Helper()
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)
	for name, content := range files {
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	return io.NopCloser(buf)
}

func TestAI_BuildkitStapel_DataArchives(t *testing.T) {
	localMounts := map[string]fsutil.FS{}
	state, cleanup, err := applyStapelDataArchives(context.Background(), llb.Scratch(), []DataArchiveSpec{
		{Archive: makeTestTar(t, map[string]string{"file.txt": "data"}), Type: DirectoryArchive, To: "/app"},
		{Archive: makeTestTar(t, map[string]string{"binary": "bin"}), Type: FileArchive, To: "/usr/bin/binary"},
	}, localMounts, testPlatform)
	defer cleanup()
	require.NoError(t, err)

	assert.Len(t, localMounts, 2)
	assert.Contains(t, localMounts, "data-archive-0")
	assert.Contains(t, localMounts, "data-archive-1")

	ops := marshalStapelOps(t, state)

	localCount := 0
	for _, op := range ops {
		if src := op.GetSource(); src != nil && (src.Identifier == "local://data-archive-0" || src.Identifier == "local://data-archive-1") {
			localCount++
		}
	}
	assert.Equal(t, 2, localCount)

	actions := stapelFileActions(t, ops)
	var copyDests []string
	var rmPaths []string
	for _, action := range actions {
		if c := action.GetCopy(); c != nil {
			copyDests = append(copyDests, c.Dest)
		}
		if rm := action.GetRm(); rm != nil {
			rmPaths = append(rmPaths, rm.Path)
		}
	}
	assert.Contains(t, copyDests, "/app")
	assert.Contains(t, copyDests, "/usr/bin")
	assert.Contains(t, rmPaths, "/usr/bin/binary")
}

func TestAI_BuildkitStapel_RemoveData(t *testing.T) {
	state := applyStapelRemoveData(llb.Scratch(), []RemoveDataSpec{
		{Type: RemoveExactPath, Paths: []string{"/exact"}},
		{Type: RemoveExactPathWithEmptyParentDirs, Paths: []string{"/parent/file"}},
		{Type: RemoveInsidePath, Paths: []string{"/inside"}},
	}, testPlatform)

	var rmPaths []string
	for _, action := range stapelFileActions(t, marshalStapelOps(t, state)) {
		rm := action.GetRm()
		require.NotNil(t, rm)
		assert.True(t, rm.AllowNotFound)
		rmPaths = append(rmPaths, rm.Path)
	}
	assert.Equal(t, []string{"/exact", "/parent/file", "/inside/*"}, rmPaths)
}

func TestAI_BuildkitStapel_Commands(t *testing.T) {
	opts := BuildStapelStageOptions{
		Commands:      []string{"echo hello"},
		Envs:          map[string]string{"CONFIG_ENV": "v1"},
		BuildTimeEnvs: map[string]string{"BUILD_ENV": "v2"},
		Network:       "none",
		BuildVolumes:  []string{"/host/build_dir:/container/build_dir"},
	}

	state, session, err := applyStapelCommands(context.Background(), llb.Scratch(), opts, testPlatform)
	require.NoError(t, err)
	assert.False(t, session.sshAgentSockUsed)

	ops := marshalStapelOps(t, state)
	var exec *pb.ExecOp
	for _, op := range ops {
		if e := op.GetExec(); e != nil {
			exec = e
		}
	}
	require.NotNil(t, exec)

	assert.Equal(t, []string{"sh", "/.werf/script.sh"}, exec.Meta.Args)
	assert.Equal(t, "0:0", exec.Meta.User)
	assert.Equal(t, "/", exec.Meta.Cwd)
	assert.Contains(t, exec.Meta.Env, "CONFIG_ENV=v1")
	assert.Contains(t, exec.Meta.Env, "BUILD_ENV=v2")
	assert.Equal(t, pb.NetMode_NONE, exec.Network)

	mountsByDest := map[string]*pb.Mount{}
	for _, m := range exec.Mounts {
		mountsByDest[m.Dest] = m
	}

	require.Contains(t, mountsByDest, "/.werf/script.sh")
	assert.True(t, mountsByDest["/.werf/script.sh"].Readonly)

	require.Contains(t, mountsByDest, "/container/build_dir")
	cacheMount := mountsByDest["/container/build_dir"]
	assert.Equal(t, pb.MountType_CACHE, cacheMount.MountType)
	assert.Equal(t, "/host/build_dir", cacheMount.CacheOpt.ID)
	assert.Equal(t, pb.CacheSharingOpt_SHARED, cacheMount.CacheOpt.Sharing)
}

func TestAI_BuildkitStapel_CommandsSSHSocket(t *testing.T) {
	opts := BuildStapelStageOptions{
		Commands:      []string{"git clone ssh://example.com/repo"},
		BuildTimeEnvs: map[string]string{"SSH_AUTH_SOCK": "/.werf/ssh-auth-sock"},
		BuildVolumes:  []string{"/host/agent.sock:/.werf/ssh-auth-sock"},
	}

	state, session, err := applyStapelCommands(context.Background(), llb.Scratch(), opts, testPlatform)
	require.NoError(t, err)
	assert.True(t, session.sshAgentSockUsed)

	ops := marshalStapelOps(t, state)
	var exec *pb.ExecOp
	for _, op := range ops {
		if e := op.GetExec(); e != nil {
			exec = e
		}
	}
	require.NotNil(t, exec)

	var sshMount *pb.Mount
	for _, m := range exec.Mounts {
		if m.MountType == pb.MountType_SSH {
			sshMount = m
		}
	}
	require.NotNil(t, sshMount)
	assert.Equal(t, "/.werf/ssh-auth-sock", sshMount.Dest)
}

func TestAI_BuildkitStapel_CommandsSecretFileVolume(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(secretFile, []byte("value"), 0o400))

	opts := BuildStapelStageOptions{
		Commands:     []string{"cat /run/secrets/mysecret"},
		BuildVolumes: []string{secretFile + ":/run/secrets/mysecret:ro"},
	}

	state, session, err := applyStapelCommands(context.Background(), llb.Scratch(), opts, testPlatform)
	require.NoError(t, err)

	require.Len(t, session.secretSources, 1)
	assert.Equal(t, "run-secrets-mysecret", session.secretSources[0].ID)
	assert.Equal(t, secretFile, session.secretSources[0].FilePath)

	ops := marshalStapelOps(t, state)
	var exec *pb.ExecOp
	for _, op := range ops {
		if e := op.GetExec(); e != nil {
			exec = e
		}
	}
	require.NotNil(t, exec)

	var secretMount *pb.Mount
	for _, m := range exec.Mounts {
		if m.MountType == pb.MountType_SECRET {
			secretMount = m
		}
	}
	require.NotNil(t, secretMount)
	assert.Equal(t, "/run/secrets/mysecret", secretMount.Dest)
	assert.Equal(t, "run-secrets-mysecret", secretMount.SecretOpt.ID)
}

func TestAI_BuildkitStapel_CommandsReadonlyVolume(t *testing.T) {
	opts := BuildStapelStageOptions{
		Commands:     []string{"ls /mnt"},
		BuildVolumes: []string{"/host/dir:/mnt:ro"},
	}

	state, _, err := applyStapelCommands(context.Background(), llb.Scratch(), opts, testPlatform)
	require.NoError(t, err)

	ops := marshalStapelOps(t, state)
	var exec *pb.ExecOp
	for _, op := range ops {
		if e := op.GetExec(); e != nil {
			exec = e
		}
	}
	require.NotNil(t, exec)

	var cacheMount *pb.Mount
	for _, m := range exec.Mounts {
		if m.Dest == "/mnt" {
			cacheMount = m
		}
	}
	require.NotNil(t, cacheMount)
	assert.Equal(t, pb.MountType_CACHE, cacheMount.MountType)
	assert.True(t, cacheMount.Readonly)
}

func TestAI_BuildkitStapel_ImageConfig(t *testing.T) {
	img := &dockerspec.DockerOCIImage{}
	opts := BuildStapelStageOptions{
		Labels:      []string{"werf-stage-content-digest=abc123", "other=value"},
		Volumes:     []string{"/data"},
		Expose:      []string{"8080", "9090/udp"},
		Envs:        map[string]string{"APP_ENV": "production"},
		Cmd:         []string{"server", "run"},
		Entrypoint:  []string{"/bin/app"},
		User:        "app:app",
		Workdir:     "/app",
		Healthcheck: `--interval=30s CMD curl -f http://localhost/`,
	}

	require.NoError(t, applyStapelImageConfig(img, opts))

	assert.Equal(t, "abc123", img.Config.Labels["werf-stage-content-digest"])
	assert.Equal(t, "value", img.Config.Labels["other"])
	assert.Contains(t, img.Config.Volumes, "/data")
	assert.Contains(t, img.Config.ExposedPorts, "8080/tcp")
	assert.Contains(t, img.Config.ExposedPorts, "9090/udp")
	assert.Contains(t, img.Config.Env, "APP_ENV=production")
	assert.Equal(t, []string{"server", "run"}, img.Config.Cmd)
	assert.Equal(t, []string{"/bin/app"}, img.Config.Entrypoint)
	assert.Equal(t, "app:app", img.Config.User)
	assert.Equal(t, "/app", img.Config.WorkingDir)
	require.NotNil(t, img.Config.Healthcheck)
	assert.Equal(t, []string{"CMD-SHELL", "curl -f http://localhost/"}, img.Config.Healthcheck.Test)

	assert.Error(t, applyStapelImageConfig(img, BuildStapelStageOptions{Labels: []string{"invalid-label"}}))
}
