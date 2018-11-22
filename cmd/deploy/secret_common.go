package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/kubernetes/pkg/util/file"
)

type secretGenerateOptions struct {
	FilePath       string `json:"file_path"`
	OutputFilePath string `json:"output_file_path"`
	Values         bool   `json:"values"`
}

func readFileData(filePath string) ([]byte, error) {
	if exist, err := file.FileExists(filePath); err != nil {
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

func readStdin() ([]byte, error) {
	var data []byte
	var err error

	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Printf("Enter secret: ")
		data, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return nil, err
		}
	} else {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func saveGeneratedData(filePath string, data []byte, options secretGenerateOptions) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(options.OutputFilePath, data, 0644); err != nil {
		return err
	}

	return nil
}
