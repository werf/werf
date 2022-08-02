package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/moby/term"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/werf/pkg/util"
)

type GenerateOptions struct {
	FilePath       string
	OutputFilePath string
	Values         bool
}

func ReadFileData(filePath string) ([]byte, error) {
	if exist, err := util.FileExists(filePath); err != nil {
		return nil, err
	} else if !exist {
		absFilePath, err := filepath.Abs(filePath)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("secret file %q not found", absFilePath)
	}

	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return fileData, err
}

func InputFromInteractiveStdin(prompt string) ([]byte, error) {
	var data []byte
	var err error

	isStdoutTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	if isStdoutTerminal {
		fmt.Printf(logboek.Colorize(style.Highlight(), prompt))
	}

	prepareTerminal := func() (func() error, error) {
		state, err := term.SetRawTerminal(os.Stdin.Fd())
		if err != nil {
			return nil, fmt.Errorf("unable to put terminal into raw mode: %w", err)
		}

		restored := false

		return func() error {
			if restored {
				return nil
			}
			if err := term.RestoreTerminal(os.Stdin.Fd(), state); err != nil {
				return err
			}
			restored = true
			return nil
		}, nil
	}

	restoreTerminal, err := prepareTerminal()
	if err != nil {
		return nil, err
	}
	defer restoreTerminal()

	data, err = terminal.ReadPassword(int(os.Stdin.Fd()))

	if err := restoreTerminal(); err != nil {
		return nil, fmt.Errorf("unable to restore terminal: %w", err)
	}

	if isStdoutTerminal {
		fmt.Println()
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

func InputFromStdin() ([]byte, error) {
	var data []byte
	var err error

	data, err = ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func SaveGeneratedData(filePath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filePath, data, 0o644); err != nil {
		return err
	}

	return nil
}
