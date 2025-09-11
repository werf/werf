package true_git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/path_matcher"
	werfExec "github.com/werf/werf/v2/pkg/werf/exec"
)

type PatchOptions struct {
	PathScope            string // Determines the directory that will get into the result (similar to <pathspec> in the git commands).
	PathMatcher          path_matcher.PathMatcher
	FromCommit, ToCommit string
	FileRenames          map[string]string // Files to rename during patching. Git repo relative paths of original files as keys, new filenames (without base path) as values.

	// TODO: maybe add --path-status option for git, so that created patch will be only presented by the descriptor without content-diff

	WithEntireFileContext bool
	WithBinary            bool
}

func (opts PatchOptions) ID() string {
	var renamedOldFilePaths, renamedNewFileNames []string
	for renamedOldFilePath, renamedNewFileName := range opts.FileRenames {
		renamedOldFilePaths = append(renamedOldFilePaths, renamedOldFilePath)
		renamedNewFileNames = append(renamedNewFileNames, renamedNewFileName)
	}

	return util.Sha256Hash(
		append(
			append(renamedOldFilePaths, renamedNewFileNames...),
			opts.FromCommit,
			opts.ToCommit,
			opts.PathScope,
			opts.PathMatcher.ID(),
			fmt.Sprint(opts.WithBinary),
			fmt.Sprint(opts.WithEntireFileContext),
		)...,
	)
}

type PatchDescriptor struct {
	Paths         []string
	BinaryPaths   []string
	PathsToRemove []string
}

func Patch(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, withSubmodules bool, opts PatchOptions) (*PatchDescriptor, error) {
	var res *PatchDescriptor

	workTreeDir := workTreeCacheDir

	if workTreePoolLimit != "" {
		poolSize, err := strconv.Atoi(workTreePoolLimit)
		if err != nil {
			return nil, fmt.Errorf("invalid WERF_GIT_WORK_TREE_POOL_LIMIT value %s: %w", workTreePoolLimit, err)
		}

		workTreePool, err := GetWorkTreePool(workTreeCacheDir, poolSize)
		if err != nil {
			return nil, fmt.Errorf("unable to create worktree pool: %w", err)
		}

		slot, wt, err := workTreePool.Acquire(ctx)
		if err != nil {
			return nil, err
		}
		defer workTreePool.Release(slot)

		workTreeDir = wt

	}
	err := withWorkTreeCacheLock(ctx, workTreeDir, func() error {
		writePatchRes, err := writePatch(ctx, out, gitDir, workTreeDir, withSubmodules, opts)
		res = writePatchRes
		return err
	})

	return res, err
}

func debugPatch() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_PATCH") == "1"
}

func writePatch(ctx context.Context, out io.Writer, gitDir, workTreeCacheDir string, withSubmodules bool, opts PatchOptions) (*PatchDescriptor, error) {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return nil, fmt.Errorf("bad git dir %s: %w", gitDir, err)
	}

	workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
	if err != nil {
		return nil, fmt.Errorf("bad work tree cache dir %s: %w", workTreeCacheDir, err)
	}

	if withSubmodules && workTreeCacheDir == "" {
		return nil, fmt.Errorf("provide work tree cache directory to enable submodules!")
	}

	commonGitOpts := append(getCommonGitOptions(),
		"-c", "diff.renames=false",
		"-c", "core.quotePath=false",
	)
	if opts.WithEntireFileContext {
		commonGitOpts = append(commonGitOpts, "-c", "diff.context=999999999")
	}

	diffOpts := []string{"--full-index"}
	if withSubmodules {
		diffOpts = append(diffOpts, "--submodule=diff")
	} else {
		diffOpts = append(diffOpts, "--submodule=log")
	}
	if opts.WithBinary {
		diffOpts = append(diffOpts, "--binary")
	}

	var cmd *exec.Cmd

	if withSubmodules {
		workTreeDir, err := prepareWorkTree(ctx, gitDir, workTreeCacheDir, opts.ToCommit, withSubmodules)
		if err != nil {
			return nil, fmt.Errorf("cannot prepare work tree in cache %s for commit %s: %w", workTreeCacheDir, opts.ToCommit, err)
		}

		gitArgs := commonGitOpts
		gitArgs = append(gitArgs, "-C", workTreeDir)
		gitArgs = append(gitArgs, "diff")
		gitArgs = append(gitArgs, diffOpts...)
		gitArgs = append(gitArgs, opts.FromCommit, opts.ToCommit)

		if debugPatch() {
			fmt.Printf("# git %s\n", strings.Join(gitArgs, " "))
		}

		cmd = werfExec.CommandContextCancellation(ctx, "git", gitArgs...)

		cmd.Dir = workTreeDir // required for `git diff` with submodules
	} else {
		gitArgs := commonGitOpts
		gitArgs = append(gitArgs, "-C", gitDir)
		gitArgs = append(gitArgs, "diff")
		gitArgs = append(gitArgs, diffOpts...)
		gitArgs = append(gitArgs, opts.FromCommit, opts.ToCommit)

		if debugPatch() {
			fmt.Printf("# git %s\n", strings.Join(gitArgs, " "))
		}

		cmd = werfExec.CommandContextCancellation(ctx, "git", gitArgs...)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating git diff stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating git diff stderr pipe: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting git diff: %w", err)
	}

	if debugPatch() {
		out = io.MultiWriter(out, os.Stdout)
	}

	p := makeDiffParser(out, opts.PathScope, opts.PathMatcher, opts.FileRenames)

	const stdoutPipeType = "stdout"
	const stderrPipeType = "stderr"

	stdoutCh := newGeneratorFromStdPipe(1, stdoutPipe, stdoutPipeType)
	stderrCh := newGeneratorFromStdPipe(1, stderrPipe, stderrPipeType)

	combinedOutputCh := lo.FanIn(1, stdoutCh, stderrCh)

	for stdItem := range combinedOutputCh {
		if stdItem.Err != nil {
			return nil, fmt.Errorf("error getting git diff output: %w\nunrecognized output:\n%s", err, p.UnrecognizedCapture.String())
		}

		switch stdItem.Type {
		case stdoutPipeType:
			if err = p.HandleStdout(stdItem.Data); err != nil {
				return nil, err
			}
		case stderrPipeType:
			if err = p.HandleStderr(stdItem.Data); err != nil {
				return nil, err
			}
		}
	}

	if err = cmd.Wait(); err != nil {
		werfExec.TerminateIfCanceled(ctx)
		return nil, fmt.Errorf("git diff error: %w\nunrecognized output:\n%s", err, p.UnrecognizedCapture.String())
	}

	desc := &PatchDescriptor{
		Paths:         p.Paths,
		BinaryPaths:   p.BinaryPaths,
		PathsToRemove: p.PathsToRemove,
	}

	if debugPatch() {
		fmt.Printf("Patch paths count is %d, binary paths count is %d\n", len(desc.Paths), len(desc.BinaryPaths))
		for _, path := range desc.Paths {
			fmt.Printf("Patch path %s\n", path)
		}
		for _, path := range desc.BinaryPaths {
			fmt.Printf("Binary patch path %s\n", path)
		}
	}

	return desc, nil
}

func consumePipeOutput(pipe io.ReadCloser, handleChunk func(data []byte) error) error {
	chunkBuf := make([]byte, 1024*64)

	for {
		n, err := pipe.Read(chunkBuf)
		if n > 0 {
			handleErr := handleChunk(chunkBuf[:n])
			if handleErr != nil {
				return handleErr
			}
		}

		if err == io.EOF {
			err := pipe.Close()
			if err != nil {
				return fmt.Errorf("error closing pipe: %w", err)
			}

			return nil
		}

		if err != nil {
			return fmt.Errorf("error reading pipe: %w", err)
		}
	}
}

type stdPipeItem struct {
	Type string
	Err  error
	Data []byte
}

func newGeneratorFromStdPipe(bufferSize int, rc io.ReadCloser, stdType string) <-chan stdPipeItem {
	return lo.Generator(bufferSize, func(yield func(stdPipeItem)) {
		err := consumePipeOutput(rc, func(data []byte) error {
			yield(stdPipeItem{Type: stdType, Err: nil, Data: bytes.Clone(data)})
			return nil
		})
		if err != nil {
			yield(stdPipeItem{Type: stdType, Err: err, Data: nil})
		}
	})
}
