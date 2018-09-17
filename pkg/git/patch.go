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

type PatchDescriptor struct {
	IsEmpty bool
}

func Patch(out io.Writer, repoPath string, opts DiffOptions) (*PatchDescriptor, error) {
	// TODO: Check repoPath exists and valid, run `git -C repoPath log -1`.
	// TODO: Do not accept anything except commit-id hash.
	// TODO: What about non-existing commits? For now go-git checks commits.

	submoduleOpt := "--submodule=log"
	if opts.WithSubmodules {
		err := checkSubmoduleConstraint()
		if err != nil {
			return nil, err
		}
		submoduleOpt = "--submodule=diff"
	}

	// TODO: Maybe use git-dir + bare always.
	// TODO: For now -C is the compatible with go-git solution.
	cmd := exec.Command(
		"git", "-C", repoPath,
		"diff", opts.FromCommit, opts.ToCommit,
		"--binary", "--no-renames", submoduleOpt,
	)

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
		IsEmpty: (p.OutLines == 0),
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
