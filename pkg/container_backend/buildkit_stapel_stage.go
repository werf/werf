package container_backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/sourceresolver"
	bkinstructions "github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/solver/pb"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/tonistiigi/fsutil"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v2/pkg/buildkit"
	"github.com/werf/werf/v2/pkg/ssh_agent"
)

func (backend *BuildkitBackend) BuildStapelStage(ctx context.Context, baseImage string, opts BuildStapelStageOptions) (string, error) {
	repo, err := backend.getStagesStorageRepo()
	if err != nil {
		return "", err
	}

	cl, err := backend.getClient(ctx)
	if err != nil {
		return "", err
	}

	platform, err := backend.parsePlatform(opts.TargetPlatform)
	if err != nil {
		return "", err
	}

	resolver := buildkit.NewImageMetaResolver(platform)
	_, baseDigest, baseConfig, err := resolver.ResolveImageConfig(ctx, baseImage, sourceresolver.Opt{
		ImageOpt: &sourceresolver.ResolveImageOpt{Platform: platform},
	})
	if err != nil {
		return "", fmt.Errorf("resolve base image %q: %w", baseImage, err)
	}

	img := &dockerspec.DockerOCIImage{}
	if err := json.Unmarshal(baseConfig, img); err != nil {
		return "", fmt.Errorf("unmarshal base image %q config: %w", baseImage, err)
	}

	state := llb.Image(fmt.Sprintf("%s@%s", strings.Split(baseImage, "@")[0], baseDigest), llb.Platform(*platform))
	state, err = state.WithImageConfig(baseConfig)
	if err != nil {
		return "", fmt.Errorf("apply base image %q config to llb state: %w", baseImage, err)
	}

	localMounts := map[string]fsutil.FS{}

	state, err = applyStapelDependencyImports(state, opts.DependencyImportSpecs, *platform)
	if err != nil {
		return "", err
	}

	state, cleanupArchives, err := applyStapelDataArchives(ctx, state, opts.DataArchiveSpecs, localMounts, *platform)
	defer cleanupArchives()
	if err != nil {
		return "", err
	}

	state = applyStapelRemoveData(state, opts.RemoveDataSpecs, *platform)

	var sshAgentSockUsed bool
	if len(opts.Commands) > 0 {
		state, sshAgentSockUsed, err = applyStapelCommands(ctx, state, opts, *platform)
		if err != nil {
			return "", err
		}
	}

	if err := applyStapelImageConfig(img, opts); err != nil {
		return "", err
	}

	def, err := state.Marshal(ctx)
	if err != nil {
		return "", fmt.Errorf("marshal llb state: %w", err)
	}

	imageConfig, err := json.Marshal(img)
	if err != nil {
		return "", fmt.Errorf("marshal image config: %w", err)
	}

	var sshSpec string
	if sshAgentSockUsed {
		sshSpec = "default"
	}
	attachables, err := backend.getSessionAttachables(sshSpec, nil)
	if err != nil {
		return "", err
	}

	builtID, err := buildkit.Solve(ctx, cl, def, buildkit.SolveOptions{
		Repo:        repo,
		ImageConfig: imageConfig,
		LocalMounts: localMounts,
		Session:     attachables,
	})
	if err != nil {
		return "", fmt.Errorf("build stapel stage: %w", err)
	}

	return builtID, nil
}

func applyStapelDependencyImports(state llb.State, imports []DependencyImportSpec, platform ocispecs.Platform) (llb.State, error) {
	for _, imp := range imports {
		copyInfo := &llb.CopyInfo{
			CopyDirContentsOnly: true,
			CreateDestPath:      true,
			AllowWildcard:       true,
			AllowEmptyWildcard:  true,
			IncludePatterns:     imp.IncludePaths,
			ExcludePatterns:     imp.ExcludePaths,
			ChownOpt:            makeChownOpt(imp.Owner, imp.Group),
		}
		src := llb.Image(imp.ImageName, llb.Platform(platform))
		state = state.File(
			llb.Copy(src, imp.FromPath, imp.ToPath, copyInfo),
			llb.Platform(platform),
			llb.WithCustomNamef("[import] %s: %s -> %s", imp.ImageName, imp.FromPath, imp.ToPath),
		)
	}
	return state, nil
}

func applyStapelDataArchives(ctx context.Context, state llb.State, archives []DataArchiveSpec, localMounts map[string]fsutil.FS, platform ocispecs.Platform) (llb.State, func(), error) {
	var tmpDirs []string
	cleanup := func() {
		for _, dir := range tmpDirs {
			os.RemoveAll(dir)
		}
	}

	for i, archive := range archives {
		tmpDir, err := os.MkdirTemp("", "werf-buildkit-data-archive-*")
		if err != nil {
			return state, cleanup, fmt.Errorf("create tmp dir for data archive: %w", err)
		}
		tmpDirs = append(tmpDirs, tmpDir)

		if err := extractTarWithChown(archive.Archive, tmpDir, nil, nil); err != nil {
			return state, cleanup, fmt.Errorf("unable to extract data archive into %s: %w", archive.To, err)
		}
		if err := archive.Archive.Close(); err != nil {
			return state, cleanup, fmt.Errorf("error closing archive data stream: %w", err)
		}

		localName := fmt.Sprintf("data-archive-%d", i)
		localFS, err := fsutil.NewFS(tmpDir)
		if err != nil {
			return state, cleanup, fmt.Errorf("create fs for data archive %q: %w", tmpDir, err)
		}
		localMounts[localName] = localFS

		local := llb.Local(localName)

		var destPath string
		switch archive.Type {
		case DirectoryArchive:
			destPath = archive.To
		case FileArchive:
			destPath = path.Dir(archive.To)
			state = state.File(
				llb.Rm(archive.To, llb.WithAllowNotFound(true)),
				llb.Platform(platform),
				llb.WithCustomNamef("[data archive] cleanup %s", archive.To),
			)
		default:
			return state, cleanup, fmt.Errorf("unknown archive type %q", archive.Type)
		}

		copyInfo := &llb.CopyInfo{
			CopyDirContentsOnly: true,
			CreateDestPath:      true,
			ChownOpt:            makeChownOpt(archive.Owner, archive.Group),
		}
		state = state.File(
			llb.Copy(local, "/", destPath, copyInfo),
			llb.Platform(platform),
			llb.WithCustomNamef("[data archive] extract into %s", archive.To),
		)
	}

	return state, cleanup, nil
}

func applyStapelRemoveData(state llb.State, removeData []RemoveDataSpec, platform ocispecs.Platform) llb.State {
	for _, spec := range removeData {
		for _, p := range spec.Paths {
			var rmPath string
			switch spec.Type {
			case RemoveInsidePath:
				rmPath = path.Join(p, "*")
			default:
				// NOTE: RemoveExactPathWithEmptyParentDirs behaves as RemoveExactPath: emptied
				// parent dirs are kept in the image, unlike the buildah backend which pruned them.
				rmPath = p
			}
			state = state.File(
				llb.Rm(rmPath, llb.WithAllowNotFound(true), llb.WithAllowWildcard(true)),
				llb.Platform(platform),
				llb.WithCustomNamef("[cleanup] remove %s", rmPath),
			)
		}
	}
	return state
}

func applyStapelCommands(ctx context.Context, state llb.State, opts BuildStapelStageOptions, platform ocispecs.Platform) (llb.State, bool, error) {
	scriptContent := makeScript(opts.Commands, logboek.Context(ctx).IsAcceptedLevel(level.Info))
	destScriptPath := "/.werf/script.sh"

	scriptState := llb.Scratch().File(
		llb.Mkfile("/script.sh", 0o555, scriptContent),
		llb.Platform(platform),
	)

	runOpts := []llb.RunOption{
		llb.Args([]string{"sh", destScriptPath}),
		llb.User("0:0"),
		llb.Dir("/"),
		llb.WithCustomName("[commands] run assembly instructions"),
		llb.AddMount(destScriptPath, scriptState, llb.SourcePath("/script.sh"), llb.Readonly),
	}

	mergedEnvs := make(map[string]string, len(opts.Envs)+len(opts.BuildTimeEnvs))
	for k, v := range opts.Envs {
		mergedEnvs[k] = v
	}
	for k, v := range opts.BuildTimeEnvs {
		mergedEnvs[k] = v
	}
	for k, v := range mergedEnvs {
		runOpts = append(runOpts, llb.AddEnv(k, v))
	}

	netMode, err := buildkit.ParseNetMode(opts.Network)
	if err != nil {
		return state, false, err
	}
	if netMode != pb.NetMode_UNSET {
		runOpts = append(runOpts, llb.Network(netMode))
	}

	sshSockTarget := opts.BuildTimeEnvs[ssh_agent.SSHAuthSockEnv]

	var sshAgentSockUsed bool
	for _, volume := range opts.BuildVolumes {
		from, to, _, err := parseVolume(volume)
		if err != nil {
			return state, false, fmt.Errorf("invalid volume %q: %w", volume, err)
		}

		if sshSockTarget != "" && to == sshSockTarget {
			sshAgentSockUsed = true
			runOpts = append(runOpts, llb.AddSSHSocket(llb.SSHSocketTarget(to)))
			continue
		}

		// Host bind mounts are not possible with a remote buildkitd: host-path-keyed shared
		// cache mounts preserve the data-reuse semantics, but the data lives in the buildkitd
		// cache instead of the host directory.
		runOpts = append(runOpts, llb.AddMount(to, llb.Scratch(), llb.AsPersistentCacheDir(from, llb.CacheMountShared)))
	}

	return state.Run(runOpts...).Root(), sshAgentSockUsed, nil
}

func applyStapelImageConfig(img *dockerspec.DockerOCIImage, opts BuildStapelStageOptions) error {
	if len(opts.Labels) > 0 && img.Config.Labels == nil {
		img.Config.Labels = map[string]string{}
	}
	for _, label := range opts.Labels {
		key, value, ok := strings.Cut(label, "=")
		if !ok {
			return fmt.Errorf("invalid label %q given, expected string in the key=value format", label)
		}
		img.Config.Labels[key] = value
	}

	if len(opts.Volumes) > 0 && img.Config.Volumes == nil {
		img.Config.Volumes = map[string]struct{}{}
	}
	for _, volume := range opts.Volumes {
		img.Config.Volumes[volume] = struct{}{}
	}

	if len(opts.Expose) > 0 && img.Config.ExposedPorts == nil {
		img.Config.ExposedPorts = map[string]struct{}{}
	}
	for _, expose := range opts.Expose {
		if !strings.Contains(expose, "/") {
			expose += "/tcp"
		}
		img.Config.ExposedPorts[expose] = struct{}{}
	}

	for key, value := range opts.Envs {
		img.Config.Env = setImageConfigEnv(img.Config.Env, key, value)
	}

	if len(opts.Cmd) > 0 {
		img.Config.Cmd = opts.Cmd
	}
	if len(opts.Entrypoint) > 0 {
		img.Config.Entrypoint = opts.Entrypoint
	}
	if opts.User != "" {
		img.Config.User = opts.User
	}
	if opts.Workdir != "" {
		img.Config.WorkingDir = opts.Workdir
	}

	if opts.Healthcheck != "" {
		healthcheck, err := parseHealthcheck(opts.Healthcheck)
		if err != nil {
			return fmt.Errorf("unable to parse healthcheck %q: %w", opts.Healthcheck, err)
		}
		img.Config.Healthcheck = healthcheck
	}

	return nil
}

func parseHealthcheck(healthcheck string) (*dockerspec.HealthcheckConfig, error) {
	dockerfile, err := parser.Parse(bytes.NewBufferString(fmt.Sprintf("HEALTHCHECK %s", healthcheck)))
	if err != nil {
		return nil, fmt.Errorf("unable to parse healthcheck instruction: %w", err)
	}

	var healthCheckNode *parser.Node
	for _, n := range dockerfile.AST.Children {
		if strings.ToLower(n.Value) == "healthcheck" {
			healthCheckNode = n
		}
	}
	if healthCheckNode == nil {
		return nil, fmt.Errorf("no valid healthcheck instruction found, got %q", healthcheck)
	}

	cmd, err := bkinstructions.ParseCommand(healthCheckNode)
	if err != nil {
		return nil, fmt.Errorf("cannot parse healthcheck instruction: %w", err)
	}

	return cmd.(*bkinstructions.HealthCheckCommand).Health, nil
}

func makeChownOpt(owner, group string) *llb.ChownOpt {
	if owner == "" && group == "" {
		return nil
	}

	makeUserOpt := func(nameOrID string) *llb.UserOpt {
		if nameOrID == "" {
			return nil
		}
		if id, err := strconv.Atoi(nameOrID); err == nil {
			return &llb.UserOpt{UID: id}
		}
		return &llb.UserOpt{Name: nameOrID}
	}

	return &llb.ChownOpt{
		User:  makeUserOpt(owner),
		Group: makeUserOpt(group),
	}
}

func setImageConfigEnv(env []string, key, value string) []string {
	for i, kv := range env {
		if k, _, ok := strings.Cut(kv, "="); ok && k == key {
			env[i] = key + "=" + value
			return env
		}
	}
	return append(env, key+"="+value)
}
