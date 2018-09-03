package main

import (
	"fmt"
	"io"
	"os/exec"
)

var aaa = 1

func handleStdout(data []byte) error {
	fmt.Printf("stdout:\n%s\n", data)
	aaa++

	// if aaa > 4 {
	// 	return fmt.Errorf("bad stdout: %s!\n", data)
	// }
	return nil
}

func handleStderr(data []byte) error {
	fmt.Printf("stderr:\n%s\n", data)
	return nil
}

func readChunkFromPipe(reader io.Reader, chunkBuf []byte, handleChunk func([]byte) error) (bool, error) {
	n, err := reader.Read(chunkBuf)
	if n > 0 {
		handleErr := handleChunk(chunkBuf[:n])
		if handleErr != nil {
			return false, handleErr
		}
	}

	if err == io.EOF {
		return true, nil
	}

	if err != nil {
		return false, fmt.Errorf("error reading pipe: %s", err)
	}

	return false, nil
}

func pipedProducer() error {
	cmd := exec.Command("../buffered-command-producer/main")

	var err error

	stdoutClosed := false
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %s", err)
	}

	stderrClosed := false
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %s", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %s", err)
	}

	chunkBuf := make([]byte, 5)
	for {
		if stdoutClosed && stderrClosed {
			break
		}

		if !stdoutClosed {
			isEof, err := readChunkFromPipe(stdoutPipe, chunkBuf, handleStdout)
			if err != nil {
				return err
			}

			if isEof {
				fmt.Printf("Closed stdout!\n")

				err := stdoutPipe.Close()
				if err != nil {
					return fmt.Errorf("error closing stdout pipe: %s", err)
				}
				stdoutClosed = true
			}
		}

		if !stderrClosed {
			isEof, err := readChunkFromPipe(stderrPipe, chunkBuf, handleStderr)
			if err != nil {
				return err
			}

			if isEof {
				fmt.Printf("Closed stderr!\n")

				err := stderrPipe.Close()
				if err != nil {
					return fmt.Errorf("error closing stderr pipe: %s", err)
				}
				stderrClosed = true
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for command: %s", err)
	}

	return nil
}

func main() {
	err := pipedProducer()
	if err != nil {
		panic(err)
	}
}
