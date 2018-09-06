package git

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type DiffOptions struct {
	FromCommit, ToCommit string
	PathFilter           PathFilter
	WithSubmodules       bool
}

func Diff(out io.Writer, repoPath string, opts DiffOptions) error {
	// TODO: Check repoPath exists and valid, run `git -C repoPath log -1`.
	// TODO: Do not accept anything except commit-id hash.
	// TODO: What about non-existing commits?

	submoduleOpt := "--submodule=log"
	if opts.WithSubmodules {
		submoduleOpt = "--submodule=diff"

		if !submoduleVersionConstraintObj.Check(gitVersionObj) {
			return fmt.Errorf("To use submodules install git >= %s! Your git version is %s.", MinGitVersionWithSubmodulesConstraint, GitVersion)
		}
	}

	cmd := exec.Command(
		"git", "-C", repoPath,
		"diff", opts.FromCommit, opts.ToCommit,
		"--binary", submoduleOpt,
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating git diff stdout pipe: %s", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating git diff stderr pipe: %s", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting git diff: %s", err)
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

	p := makeDiffParser(out, opts.PathFilter)

WaitForData:
	for {
		select {
		case err := <-outputErrChan:
			return fmt.Errorf("error getting git diff output: %s\nunrecognized output:\n%s", err, p.UnrecognizedCapture.String())
		case stdoutData := <-stdoutChan:
			if err := p.HandleStdout(stdoutData); err != nil {
				return err
			}
		case stderrData := <-stderrChan:
			if err := p.HandleStderr(stderrData); err != nil {
				return err
			}
		case <-doneChan:
			break WaitForData
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("git diff error: %s\nunrecognized output:\n%s", err, p.UnrecognizedCapture.String())
	}

	return nil
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
