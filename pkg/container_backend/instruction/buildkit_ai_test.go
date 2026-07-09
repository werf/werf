package instruction

import (
	"context"
	"strings"
	"testing"

	"github.com/moby/buildkit/client/llb"
	bkinstructions "github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/solver/pb"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/buildkit"
)

func newTestStage() *buildkit.DockerfileStageState {
	return &buildkit.DockerfileStageState{
		State:    llb.Scratch(),
		Image:    &dockerspec.DockerOCIImage{},
		Platform: ocispecs.Platform{OS: "linux", Architecture: "amd64"},
	}
}

func marshalOps(t *testing.T, stage *buildkit.DockerfileStageState) []*pb.Op {
	t.Helper()
	def, err := stage.State.Marshal(context.Background())
	require.NoError(t, err)

	var ops []*pb.Op
	for _, raw := range def.Def {
		op := &pb.Op{}
		require.NoError(t, op.UnmarshalVT(raw))
		ops = append(ops, op)
	}
	return ops
}

func findExecOps(ops []*pb.Op) []*pb.ExecOp {
	var res []*pb.ExecOp
	for _, op := range ops {
		if exec := op.GetExec(); exec != nil {
			res = append(res, exec)
		}
	}
	return res
}

func findFileOps(ops []*pb.Op) []*pb.FileOp {
	var res []*pb.FileOp
	for _, op := range ops {
		if file := op.GetFile(); file != nil {
			res = append(res, file)
		}
	}
	return res
}

func findSourceOps(ops []*pb.Op) []*pb.SourceOp {
	var res []*pb.SourceOp
	for _, op := range ops {
		if src := op.GetSource(); src != nil {
			res = append(res, src)
		}
	}
	return res
}

func parseRunCommand(t *testing.T, line string) *bkinstructions.RunCommand {
	t.Helper()
	parsed, err := parser.Parse(strings.NewReader(line))
	require.NoError(t, err)
	cmd, err := bkinstructions.ParseInstruction(parsed.AST.Children[0])
	require.NoError(t, err)
	run, ok := cmd.(*bkinstructions.RunCommand)
	require.True(t, ok)
	require.NoError(t, run.Expand(func(word string) (string, error) { return word, nil }))
	return run
}

func TestAI_Buildkit_ConfigInstructions(t *testing.T) {
	ctx := context.Background()
	stage := newTestStage()

	require.NoError(t, NewEnv(bkinstructions.EnvCommand{Env: []bkinstructions.KeyValuePair{{Key: "FOO", Value: "bar"}}}).ApplyBuildkit(ctx, stage))
	assert.Contains(t, stage.Image.Config.Env, "FOO=bar")

	require.NoError(t, NewLabel(bkinstructions.LabelCommand{Labels: []bkinstructions.KeyValuePair{{Key: "l1", Value: "v1"}}}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, "v1", stage.Image.Config.Labels["l1"])

	require.NoError(t, NewExpose(bkinstructions.ExposeCommand{Ports: []string{"80", "443/udp"}}).ApplyBuildkit(ctx, stage))
	assert.Contains(t, stage.Image.Config.ExposedPorts, "80/tcp")
	assert.Contains(t, stage.Image.Config.ExposedPorts, "443/udp")

	require.NoError(t, NewVolume(bkinstructions.VolumeCommand{Volumes: []string{"/data"}}).ApplyBuildkit(ctx, stage))
	assert.Contains(t, stage.Image.Config.Volumes, "/data")

	require.NoError(t, NewUser(bkinstructions.UserCommand{User: "1000:1000"}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, "1000:1000", stage.Image.Config.User)

	require.NoError(t, NewWorkdir(bkinstructions.WorkdirCommand{Path: "/app"}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, "/app", stage.Image.Config.WorkingDir)

	require.NoError(t, NewCmd(bkinstructions.CmdCommand{ShellDependantCmdLine: bkinstructions.ShellDependantCmdLine{CmdLine: []string{"server", "run"}, PrependShell: true}}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, []string{"/bin/sh", "-c", "server run"}, stage.Image.Config.Cmd)

	require.NoError(t, NewEntrypoint(bkinstructions.EntrypointCommand{ShellDependantCmdLine: bkinstructions.ShellDependantCmdLine{CmdLine: []string{"/bin/app"}}}, true).ApplyBuildkit(ctx, stage))
	assert.Equal(t, []string{"/bin/app"}, stage.Image.Config.Entrypoint)
	assert.Nil(t, stage.Image.Config.Cmd)

	require.NoError(t, NewShell(bkinstructions.ShellCommand{Shell: []string{"/bin/bash", "-c"}}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, []string{"/bin/bash", "-c"}, stage.Image.Config.Shell)

	require.NoError(t, NewStopSignal(bkinstructions.StopSignalCommand{Signal: "SIGTERM"}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, "SIGTERM", stage.Image.Config.StopSignal)

	require.NoError(t, NewOnBuild(bkinstructions.OnbuildCommand{Expression: "RUN echo hi"}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, []string{"RUN echo hi"}, stage.Image.Config.OnBuild)

	require.NoError(t, NewMaintainer(bkinstructions.MaintainerCommand{Maintainer: "dev@werf.io"}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, "dev@werf.io", stage.Image.Author)

	healthcheck := &dockerspec.HealthcheckConfig{Test: []string{"CMD", "curl", "-f", "http://localhost/"}}
	require.NoError(t, NewHealthcheck(bkinstructions.HealthCheckCommand{Health: healthcheck}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, healthcheck, stage.Image.Config.Healthcheck)
}

func TestAI_Buildkit_Run_PrependShellAndEnvs(t *testing.T) {
	stage := newTestStage()
	run := NewRun(*parseRunCommand(t, `RUN echo hello`), []string{"MYENV=v"}, nil, "")
	require.NoError(t, run.ApplyBuildkit(context.Background(), stage))

	execOps := findExecOps(marshalOps(t, stage))
	require.Len(t, execOps, 1)
	assert.Equal(t, []string{"/bin/sh", "-c", "echo hello"}, execOps[0].Meta.Args)
	assert.Contains(t, execOps[0].Meta.Env, "MYENV=v")
}

func TestAI_Buildkit_Run_Mounts(t *testing.T) {
	stage := newTestStage()
	run := NewRun(*parseRunCommand(t, `RUN --mount=type=cache,id=gocache,target=/cache,sharing=locked --mount=type=tmpfs,target=/tmpfs --mount=type=secret,id=mysecret,target=/run/secrets/mysecret --mount=type=ssh echo hello`), nil, nil, "")
	require.NoError(t, run.ApplyBuildkit(context.Background(), stage))

	execOps := findExecOps(marshalOps(t, stage))
	require.Len(t, execOps, 1)

	mountsByDest := map[string]*pb.Mount{}
	for _, m := range execOps[0].Mounts {
		mountsByDest[m.Dest] = m
	}

	require.Contains(t, mountsByDest, "/cache")
	assert.Equal(t, pb.MountType_CACHE, mountsByDest["/cache"].MountType)
	assert.Equal(t, "gocache", mountsByDest["/cache"].CacheOpt.ID)
	assert.Equal(t, pb.CacheSharingOpt_LOCKED, mountsByDest["/cache"].CacheOpt.Sharing)

	require.Contains(t, mountsByDest, "/tmpfs")
	assert.Equal(t, pb.MountType_TMPFS, mountsByDest["/tmpfs"].MountType)

	require.Contains(t, mountsByDest, "/run/secrets/mysecret")
	assert.Equal(t, pb.MountType_SECRET, mountsByDest["/run/secrets/mysecret"].MountType)
	assert.Equal(t, "mysecret", mountsByDest["/run/secrets/mysecret"].SecretOpt.ID)

	var sshMountFound bool
	for _, m := range execOps[0].Mounts {
		if m.MountType == pb.MountType_SSH {
			sshMountFound = true
		}
	}
	assert.True(t, sshMountFound)
}

func TestAI_Buildkit_Run_NetworkAndSecurity(t *testing.T) {
	stage := newTestStage()
	run := NewRun(*parseRunCommand(t, `RUN --network=none --security=insecure echo hello`), nil, nil, "")
	require.NoError(t, run.ApplyBuildkit(context.Background(), stage))

	execOps := findExecOps(marshalOps(t, stage))
	require.Len(t, execOps, 1)
	assert.Equal(t, pb.NetMode_NONE, execOps[0].Network)
	assert.Equal(t, pb.SecurityMode_INSECURE, execOps[0].Security)
}

func TestAI_Buildkit_Copy_FromContext(t *testing.T) {
	stage := newTestStage()
	copyInstr := NewCopy(bkinstructions.CopyCommand{
		SourcesAndDest: bkinstructions.SourcesAndDest{SourcePaths: []string{"src"}, DestPath: "/dest/"},
		Chown:          "1000:1000",
		Chmod:          "755",
	})
	require.NoError(t, copyInstr.ApplyBuildkit(context.Background(), stage))
	assert.True(t, stage.UsesContext)

	ops := marshalOps(t, stage)
	fileOps := findFileOps(ops)
	require.Len(t, fileOps, 1)
	require.Len(t, fileOps[0].Actions, 1)

	copyAction := fileOps[0].Actions[0].GetCopy()
	require.NotNil(t, copyAction)
	assert.Equal(t, "/src", copyAction.Src)
	assert.Equal(t, "/dest/", copyAction.Dest)
	assert.True(t, copyAction.DirCopyContents)
	assert.True(t, copyAction.FollowSymlink)
	assert.True(t, copyAction.CreateDestPath)
	assert.False(t, copyAction.AttemptUnpackDockerCompatibility)
	require.NotNil(t, copyAction.Owner)
	assert.EqualValues(t, 0o755, copyAction.Mode)

	var localFound bool
	for _, src := range findSourceOps(ops) {
		if src.Identifier == "local://context" {
			localFound = true
		}
	}
	assert.True(t, localFound)
}

func TestAI_Buildkit_Copy_FromImage(t *testing.T) {
	stage := newTestStage()
	copyInstr := NewCopy(bkinstructions.CopyCommand{
		SourcesAndDest: bkinstructions.SourcesAndDest{SourcePaths: []string{"/bin/app"}, DestPath: "/usr/local/bin/app"},
		From:           "registry.example.com/project:dep-image",
	})
	require.NoError(t, copyInstr.ApplyBuildkit(context.Background(), stage))
	assert.False(t, stage.UsesContext)

	var imageFound bool
	for _, src := range findSourceOps(marshalOps(t, stage)) {
		if strings.Contains(src.Identifier, "registry.example.com/project:dep-image") {
			imageFound = true
		}
	}
	assert.True(t, imageFound)
}

func TestAI_Buildkit_Add_HTTPAndUnpack(t *testing.T) {
	stage := newTestStage()
	addInstr := NewAdd(bkinstructions.AddCommand{
		SourcesAndDest: bkinstructions.SourcesAndDest{SourcePaths: []string{"https://example.com/archive.tar.gz", "local.tar"}, DestPath: "/dest/"},
	})
	require.NoError(t, addInstr.ApplyBuildkit(context.Background(), stage))

	ops := marshalOps(t, stage)

	var httpFound bool
	for _, src := range findSourceOps(ops) {
		if strings.HasPrefix(src.Identifier, "https://example.com/archive.tar.gz") {
			httpFound = true
		}
	}
	assert.True(t, httpFound)

	fileOps := findFileOps(ops)
	require.Len(t, fileOps, 1)
	require.Len(t, fileOps[0].Actions, 2)

	localCopy := fileOps[0].Actions[1].GetCopy()
	require.NotNil(t, localCopy)
	assert.True(t, localCopy.AttemptUnpackDockerCompatibility)
}

func TestAI_Buildkit_Workdir_RelativePath(t *testing.T) {
	ctx := context.Background()
	stage := newTestStage()
	require.NoError(t, NewWorkdir(bkinstructions.WorkdirCommand{Path: "/app"}).ApplyBuildkit(ctx, stage))
	require.NoError(t, NewWorkdir(bkinstructions.WorkdirCommand{Path: "subdir"}).ApplyBuildkit(ctx, stage))
	assert.Equal(t, "/app/subdir", stage.Image.Config.WorkingDir)
}

func TestAI_Buildkit_Run_CollectsSecretsAndSSH(t *testing.T) {
	stage := newTestStage()
	run := NewRun(*parseRunCommand(t, `RUN echo hello`), nil, []string{"id=mysecret,src=/tmp/secret"}, "default")
	require.NoError(t, run.ApplyBuildkit(context.Background(), stage))

	assert.Equal(t, []string{"id=mysecret,src=/tmp/secret"}, stage.Secrets)
	assert.Equal(t, "default", stage.SSH)
}
