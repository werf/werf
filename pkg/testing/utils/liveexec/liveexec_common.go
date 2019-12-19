// +build !windows

package liveexec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
)

func doExecCommand(dir, binPath string, opts ExecCommandOptions, arg ...string) error {
	cmd := exec.Command(binPath, arg...)
	cmd.Dir = dir

	cmd.Env = os.Environ()
	for k, v := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	absDir, _ := filepath.Abs(dir)
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "\n[DEBUG] COMMAND in %s: %s %s\n\n", absDir, binPath, strings.Join(arg, " "))

	stdoutReadPipe, stdoutWritePipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("unable to create os pipe for stdout: %s", err)
	}

	stderrReadPipe, stderrWritePipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("unable to create os pipe for stderr: %s", err)
	}

	outputReader := io.MultiReader(stdoutReadPipe, stderrReadPipe)

	session, err := gexec.Start(cmd, stdoutWritePipe, stderrWritePipe)
	if err != nil {
		return fmt.Errorf("error starting command: %s", err)
	}

	var exitCode int

	go func() {
		<-session.Exited
		exitCode = session.ExitCode()
		// Initiate EOF for consumeOutputUntilEOF
		stdoutWritePipe.Close()
		stderrWritePipe.Close()
	}()

	lineBuf := make([]byte, 0, 4096)
	if err := consumeOutputUntilEOF(outputReader, func(data []byte) error {
		for _, b := range data {
			if b == '\n' {
				line := string(lineBuf)
				lineBuf = lineBuf[:0]

				_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "[DEBUG] OUTPUT LINE: %s\n", line)

				if opts.OutputLineHandler != nil {
					func() {
						handlerDone := false
						defer func() {
							// Cleanup process in the case of gomega panic in OutputLineHandler.
							// Current werf process may held a lock and this will lead to a deadlock in the
							// case when another werf command has been ran from AfterEach when this panic was occurred.
							//
							// Panic in OutputLineHandler and current command kill allows to fail fast
							// and give user immediate feedback of failed assertion during command execution.
							if !handlerDone {
								session.Kill()
							}
						}()
						opts.OutputLineHandler(line)
						handlerDone = true
					}()
				}

				continue
			}

			lineBuf = append(lineBuf, b)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("unable to consume command output: %s", err)
	}

	if exitCode != 0 {
		return fmt.Errorf("exit code %d", exitCode)
	}
	return nil
}

func consumeOutputUntilEOF(reader io.Reader, handleChunk func(data []byte) error) error {
	chunkBuf := make([]byte, 1024*64)

	for {
		n, err := reader.Read(chunkBuf)
		if n > 0 {
			if handleErr := handleChunk(chunkBuf[:n]); handleErr != nil {
				return handleErr
			}
		}

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("read error: %s", err)
		}
	}
}
