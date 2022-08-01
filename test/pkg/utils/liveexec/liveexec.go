package liveexec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const sessionTimeout = 60 * 15

// ExecCommandOptions is an options for ExecCommand
type ExecCommandOptions struct {
	Env               map[string]string
	OutputLineHandler func(string)
}

// ExecCommand allows handling output of executed command in realtime by CommandOptions.OutputLineHandler.
// User could set expectations on the output lines in the CommandOptions.OutputLineHandler to fail fast
// and give immediate feedback of failed assertion during command execution.
func ExecCommand(dir, binPath string, opts ExecCommandOptions, arg ...string) error {
	return doExecCommand(dir, binPath, opts, arg...)
}

func doExecCommand(dir, binPath string, opts ExecCommandOptions, arg ...string) error {
	cmd := exec.Command(binPath, arg...)
	cmd.Dir = dir

	cmd.Env = os.Environ()
	for k, v := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	absDir, err := filepath.Abs(dir)
	Ω(err).ShouldNot(HaveOccurred())
	_, _ = fmt.Fprintf(GinkgoWriter, "\n%s %s (dir: %s)\n\n", binPath, strings.Join(arg, " "), absDir)

	reader, writer, err := os.Pipe()
	Ω(err).ShouldNot(HaveOccurred(), fmt.Sprintf("unable to create os pipe: %s", err))

	multiOutWriter := io.MultiWriter(writer, GinkgoWriter)
	multiErrWriter := io.MultiWriter(writer, GinkgoWriter)

	session, err := gexec.Start(cmd, multiOutWriter, multiErrWriter)
	Ω(err).ShouldNot(HaveOccurred())
	defer session.Kill()

	var exitCode int
	go func() {
		Eventually(session, sessionTimeout).Should(gexec.Exit())
		exitCode = session.ExitCode()
		// Initiate EOF for consumeOutputUntilEOF
		_ = writer.Close()
	}()

	lineBuf := make([]byte, 0, 4096)
	err = consumeOutputUntilEOF(reader, func(data []byte) error {
		for _, b := range data {
			if b == '\n' {
				line := string(lineBuf)
				lineBuf = lineBuf[:0]

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
	})
	Ω(err).ShouldNot(HaveOccurred(), fmt.Sprintf("unable to consume command output: %s", err))

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
			return fmt.Errorf("read error: %w", err)
		}
	}
}
