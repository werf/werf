package true_git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type PatchOptions struct {
	FromCommit, ToCommit string
	PathFilter           PathFilter

	WithEntireFileContext bool
	WithBinary            bool
}

type PatchDescriptor struct {
	Paths       []string
	BinaryPaths []string
}

func PatchWithSubmodules(out io.Writer, gitDir, workTreeDir string, opts PatchOptions) (*PatchDescriptor, error) {
	return writePatch(out, gitDir, workTreeDir, true, opts)
}

func Patch(out io.Writer, gitDir string, opts PatchOptions) (*PatchDescriptor, error) {
	return writePatch(out, gitDir, "", false, opts)
}

func debugPatch() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_PATCH") == "1"
}

func writePatch(out io.Writer, gitDir, workTreeDir string, withSubmodules bool, opts PatchOptions) (*PatchDescriptor, error) {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return nil, fmt.Errorf("bad git dir `%s`: %s", gitDir, err)
	}

	workTreeDir, err = filepath.Abs(workTreeDir)
	if err != nil {
		return nil, fmt.Errorf("bad work tree dir `%s`: %s", workTreeDir, err)
	}

	if withSubmodules {
		err := checkSubmoduleConstraint()
		if err != nil {
			return nil, err
		}
	}
	if withSubmodules && workTreeDir == "" {
		return nil, fmt.Errorf("provide work tree directory to enable submodules!")
	}

	commonGitOpts := []string{
		"--git-dir", gitDir,
		"-c", "diff.renames=false",
	}
	if opts.WithEntireFileContext {
		commonGitOpts = append(commonGitOpts, "-c", "diff.context=999999999")
	}

	diffOpts := []string{}
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
		var err error

		err = switchWorkTree(gitDir, workTreeDir, opts.ToCommit)
		if err != nil {
			return nil, fmt.Errorf("cannot reset work tree `%s` to commit `%s`: %s", workTreeDir, opts.ToCommit, err)
		}

		err = deinitSubmodules(gitDir, workTreeDir)
		if err != nil {
			return nil, fmt.Errorf("cannot deinit submodules: %s", err)
		}

		err = updateSubmodules(gitDir, workTreeDir)
		if err != nil {
			return nil, fmt.Errorf("cannot update submodules: %s", err)
		}

		gitArgs := append(commonGitOpts, "--work-tree", workTreeDir)
		gitArgs = append(gitArgs, "diff")
		gitArgs = append(gitArgs, diffOpts...)
		gitArgs = append(gitArgs, opts.FromCommit, opts.ToCommit)

		if debugPatch() {
			fmt.Printf("# git %s\n", strings.Join(gitArgs, " "))
		}

		cmd = exec.Command("git", gitArgs...)

		cmd.Dir = workTreeDir // required for `git diff` with submodules
	} else {
		gitArgs := append(commonGitOpts, "diff")
		gitArgs = append(gitArgs, diffOpts...)
		gitArgs = append(gitArgs, opts.FromCommit, opts.ToCommit)

		if debugPatch() {
			fmt.Printf("# git %s\n", strings.Join(gitArgs, " "))
		}

		cmd = exec.Command("git", gitArgs...)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating git diff stdout pipe: %s", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating git diff stderr pipe: %s", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting git diff: %s", err)
	}

	outputErrChan := make(chan error)
	stdoutChan := make(chan []byte)
	stderrChan := make(chan []byte)
	doneChan := make(chan bool)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err := consumePipeOutput(stdoutPipe, func(data []byte) error {
			dataCopy := make([]byte, len(data))
			copy(dataCopy, data)

			stdoutChan <- dataCopy

			return nil
		})
		if err != nil {
			outputErrChan <- err
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err := consumePipeOutput(stderrPipe, func(data []byte) error {
			dataCopy := make([]byte, len(data))
			copy(dataCopy, data)

			stderrChan <- dataCopy

			return nil
		})
		if err != nil {
			outputErrChan <- err
		}

		wg.Done()
	}()

	go func() {
		wg.Wait()
		doneChan <- true
	}()

	if debugPatch() {
		out = io.MultiWriter(out, os.Stdout)
	}

	p := makeDiffParser(out, opts.PathFilter)

WaitForData:
	for {
		select {
		case err := <-outputErrChan:
			return nil, fmt.Errorf("error getting git diff output: %s\nunrecognized output:\n%s", err, p.UnrecognizedCapture.String())
		case stdoutData := <-stdoutChan:
			if err := p.HandleStdout(stdoutData); err != nil {
				return nil, err
			}
		case stderrData := <-stderrChan:
			if err := p.HandleStderr(stderrData); err != nil {
				return nil, err
			}
		case <-doneChan:
			break WaitForData
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git diff error: %s\nunrecognized output:\n%s", err, p.UnrecognizedCapture.String())
	}

	desc := &PatchDescriptor{
		Paths:       p.Paths,
		BinaryPaths: p.BinaryPaths,
	}

	if debugPatch() {
		fmt.Printf("Patch paths count is %d, binary paths count is %d\n", len(desc.Paths), len(desc.BinaryPaths))
		for _, path := range desc.Paths {
			fmt.Printf("Patch path `%s`\n", path)
		}
		for _, path := range desc.BinaryPaths {
			fmt.Printf("Binary patch path `%s`\n", path)
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
				return fmt.Errorf("error closing pipe: %s", err)
			}

			return nil
		}

		if err != nil {
			return fmt.Errorf("error reading pipe: %s", err)
		}
	}
}
