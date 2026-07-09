package buildkit

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/moby/buildkit/client/llb"
	bkinstructions "github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerui"
	"github.com/moby/buildkit/solver/pb"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
)

// DockerfileStageState is the mutable state of a single staged-Dockerfile stage build:
// the LLB state of the rootfs plus the image config accumulated by config instructions.
type DockerfileStageState struct {
	State    llb.State
	Image    *dockerspec.DockerOCIImage
	Platform ocispecs.Platform

	UsesContext bool
	Secrets     []string
	SSH         string
}

func (s *DockerfileStageState) ContextState() llb.State {
	return llb.Local(dockerui.DefaultLocalNameContext)
}

func (s *DockerfileStageState) AddEnv(key, value string) {
	s.State = s.State.AddEnv(key, value)
	s.Image.Config.Env = setEnv(s.Image.Config.Env, key, value)
}

func (s *DockerfileStageState) SetWorkdir(ctx context.Context, workdir string) error {
	if !path.IsAbs(workdir) {
		currentDir, err := s.State.GetDir(ctx)
		if err != nil {
			return fmt.Errorf("get current working dir: %w", err)
		}
		workdir = path.Join("/", currentDir, workdir)
	}
	s.State = s.State.Dir(workdir).File(
		llb.Mkdir(workdir, 0o755, llb.WithParents(true)),
		llb.Platform(s.Platform),
	)
	s.Image.Config.WorkingDir = workdir
	return nil
}

func (s *DockerfileStageState) SetUser(user string) {
	s.State = s.State.User(user)
	s.Image.Config.User = user
}

func (s *DockerfileStageState) WithShell(args []string, prependShell bool) []string {
	if !prependShell {
		return args
	}
	shell := s.Image.Config.Shell
	if len(shell) == 0 {
		shell = []string{"/bin/sh", "-c"}
	}
	return append(append([]string{}, shell...), strings.Join(args, " "))
}

type RunCommandOptions struct {
	Envs     []string
	Insecure bool
	Network  string
	Mounts   []*bkinstructions.Mount
}

func (s *DockerfileStageState) RunCommand(args []string, opts RunCommandOptions) error {
	runOpts := []llb.RunOption{
		llb.Args(args),
		llb.WithCustomNamef("RUN %s", strings.Join(args, " ")),
	}

	for _, env := range opts.Envs {
		key, value, ok := strings.Cut(env, "=")
		if !ok {
			return fmt.Errorf("invalid env %q given, expected string in the key=value format", env)
		}
		runOpts = append(runOpts, llb.AddEnv(key, value))
	}

	if opts.Insecure {
		runOpts = append(runOpts, llb.Security(llb.SecurityModeInsecure))
	}

	netMode, err := ParseNetMode(opts.Network)
	if err != nil {
		return err
	}
	if netMode != pb.NetMode_UNSET {
		runOpts = append(runOpts, llb.Network(netMode))
	}

	for _, mount := range opts.Mounts {
		mountOpts, err := s.runMountOptions(mount)
		if err != nil {
			return err
		}
		runOpts = append(runOpts, mountOpts...)
	}

	s.State = s.State.Run(runOpts...).Root()
	return nil
}

func (s *DockerfileStageState) runMountOptions(mount *bkinstructions.Mount) ([]llb.RunOption, error) {
	switch mount.Type {
	case bkinstructions.MountTypeBind:
		var source llb.State
		if mount.From == "" {
			source = s.ContextState()
			s.UsesContext = true
		} else {
			source = llb.Image(mount.From, llb.Platform(s.Platform))
		}

		mountOpts := []llb.MountOption{llb.SourcePath(mount.Source)}
		if mount.ReadOnly {
			mountOpts = append(mountOpts, llb.Readonly)
		}
		return []llb.RunOption{llb.AddMount(mount.Target, source, mountOpts...)}, nil
	case bkinstructions.MountTypeCache:
		sharing := llb.CacheMountShared
		switch mount.CacheSharing {
		case bkinstructions.MountSharingPrivate:
			sharing = llb.CacheMountPrivate
		case bkinstructions.MountSharingLocked:
			sharing = llb.CacheMountLocked
		}
		cacheID := mount.CacheID
		if cacheID == "" {
			cacheID = mount.Target
		}
		return []llb.RunOption{llb.AddMount(mount.Target, llb.Scratch(), llb.AsPersistentCacheDir(cacheID, sharing))}, nil
	case bkinstructions.MountTypeTmpfs:
		return []llb.RunOption{llb.AddMount(mount.Target, llb.Scratch(), llb.Tmpfs())}, nil
	case bkinstructions.MountTypeSecret:
		secretID := mount.CacheID
		if secretID == "" {
			if mount.Source != "" {
				secretID = mount.Source
			} else {
				secretID = path.Base(mount.Target)
			}
		}
		secretOpts := []llb.SecretOption{llb.SecretID(secretID)}
		if !mount.Required {
			secretOpts = append(secretOpts, llb.SecretOptional)
		}
		if mount.Mode != nil || mount.UID != nil || mount.GID != nil {
			var uid, gid int
			mode := 0o400
			if mount.UID != nil {
				uid = int(*mount.UID)
			}
			if mount.GID != nil {
				gid = int(*mount.GID)
			}
			if mount.Mode != nil {
				mode = int(*mount.Mode)
			}
			secretOpts = append(secretOpts, llb.SecretFileOpt(uid, gid, mode))
		}
		target := mount.Target
		if target == "" {
			target = "/run/secrets/" + secretID
		}
		return []llb.RunOption{llb.AddSecret(target, secretOpts...)}, nil
	case bkinstructions.MountTypeSSH:
		var sshOpts []llb.SSHOption
		if mount.CacheID != "" {
			sshOpts = append(sshOpts, llb.SSHID(mount.CacheID))
		}
		if mount.Target != "" {
			sshOpts = append(sshOpts, llb.SSHSocketTarget(mount.Target))
		}
		if !mount.Required {
			sshOpts = append(sshOpts, llb.SSHOptional)
		}
		return []llb.RunOption{llb.AddSSHSocket(sshOpts...)}, nil
	default:
		return nil, fmt.Errorf("mount type %q is not supported by buildkit backend", mount.Type)
	}
}

type CopyOptions struct {
	From  string
	Chown string
	Chmod string
	IsAdd bool
}

func (s *DockerfileStageState) Copy(ctx context.Context, params bkinstructions.SourcesAndDest, opts CopyOptions) error {
	dest, err := s.destPath(ctx, params.DestPath)
	if err != nil {
		return err
	}

	var copyOpt []llb.CopyOption
	if opts.Chown != "" {
		copyOpt = append(copyOpt, llb.WithUser(opts.Chown))
	}

	chmodOpt, err := parseChmodOpt(opts.Chmod)
	if err != nil {
		return err
	}

	var source llb.State
	if opts.From == "" {
		source = s.ContextState()
	} else {
		source = llb.Image(opts.From, llb.Platform(s.Platform))
	}

	var fileAction *llb.FileAction
	appendCopy := func(st llb.State, src string, actionOpts ...llb.CopyOption) {
		if fileAction == nil {
			fileAction = llb.Copy(st, src, dest, actionOpts...)
		} else {
			fileAction = fileAction.Copy(st, src, dest, actionOpts...)
		}
	}

	for _, src := range params.SourcePaths {
		switch {
		case isHTTPSource(src):
			if !opts.IsAdd {
				return fmt.Errorf("source can't be a URL for COPY")
			}

			filename := "__unnamed__"
			if u, err := url.Parse(src); err == nil {
				if base := path.Base(u.Path); base != "." && base != "/" {
					filename = base
				}
			}

			httpState := llb.HTTP(src, llb.Filename(filename))
			appendCopy(httpState, filename, append([]llb.CopyOption{&llb.CopyInfo{
				Mode:           chmodOpt,
				CreateDestPath: true,
			}}, copyOpt...)...)
		default:
			if opts.From == "" {
				s.UsesContext = true
			}
			appendCopy(source, path.Join("/", src), append([]llb.CopyOption{&llb.CopyInfo{
				Mode:                chmodOpt,
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
				AttemptUnpack:       opts.IsAdd,
				CreateDestPath:      true,
				AllowWildcard:       true,
				AllowEmptyWildcard:  true,
			}}, copyOpt...)...)
		}
	}

	for _, src := range params.SourceContents {
		heredocState := llb.Scratch().File(
			llb.Mkfile(path.Join("/", src.Path), 0o644, []byte(src.Data)),
			llb.Platform(s.Platform),
		)
		appendCopy(heredocState, path.Join("/", src.Path), append([]llb.CopyOption{&llb.CopyInfo{
			Mode:           chmodOpt,
			CreateDestPath: true,
		}}, copyOpt...)...)
	}

	if fileAction == nil {
		return fmt.Errorf("no sources given for copy")
	}

	s.State = s.State.File(fileAction, llb.Platform(s.Platform))
	return nil
}

func (s *DockerfileStageState) destPath(ctx context.Context, dest string) (string, error) {
	if path.IsAbs(dest) {
		return dest, nil
	}
	currentDir, err := s.State.GetDir(ctx)
	if err != nil {
		return "", fmt.Errorf("get current working dir: %w", err)
	}
	// preserve trailing slash: it makes copy treat dest as a directory
	trailingSlash := strings.HasSuffix(dest, "/")
	res := path.Join("/", currentDir, dest)
	if trailingSlash {
		res += "/"
	}
	return res, nil
}

func ParseNetMode(network string) (pb.NetMode, error) {
	switch network {
	case "", "default", "sandbox":
		return pb.NetMode_UNSET, nil
	case "host":
		return pb.NetMode_HOST, nil
	case "none":
		return pb.NetMode_NONE, nil
	default:
		return pb.NetMode_UNSET, fmt.Errorf("unsupported network mode %q for buildkit backend", network)
	}
}

func parseChmodOpt(chmod string) (*llb.ChmodOpt, error) {
	if chmod == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(chmod, 8, 32)
	if err != nil || parsed > 0o7777 {
		return nil, fmt.Errorf("invalid chmod parameter %q: it should be octal string and between 0 and 07777", chmod)
	}
	return &llb.ChmodOpt{Mode: os.FileMode(parsed)}, nil
}

func isHTTPSource(src string) bool {
	return strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")
}

func setEnv(env []string, key, value string) []string {
	for i, kv := range env {
		if k, _, ok := strings.Cut(kv, "="); ok && k == key {
			env[i] = key + "=" + value
			return env
		}
	}
	return append(env, key+"="+value)
}
