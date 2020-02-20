package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/util"
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

		return nil, fmt.Errorf("secret file '%s' not found", absFilePath)
	}

	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return fileData, err
}

func InputFromInteractiveStdin() ([]byte, error) {
	var data []byte
	var err error

	isStdoutTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	if isStdoutTerminal {
		fmt.Printf(logboek.HighlightStyle().Colorize("Enter secret: "))
	}

	data, err = terminal.ReadPassword(int(os.Stdin.Fd()))

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
	if err := os.MkdirAll(filepath.Dir(filePath), 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}
