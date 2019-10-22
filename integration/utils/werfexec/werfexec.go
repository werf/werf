package werfexec

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/onsi/gomega/gexec"
)

type CommandOptions struct {
	OutputLineHandler func(string)
}

func ExecWerfCommand(dir string, opts CommandOptions, arg ...string) error {
	cmd := exec.Command("werf", arg...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

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

	go func() {
		<-session.Exited

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

				if opts.OutputLineHandler != nil {
					fmt.Printf("[DEBUG] HANDLE OUTPUT LINE: %s\n", line)

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

	if exitCode := session.ExitCode(); exitCode != 0 {
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
