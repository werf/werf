package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/dapp/pkg/deploy"

	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/kubernetes/pkg/util/file"
)

type secretGenerateOptions struct {
	FilePath       string `json:"file_path"`
	OutputFilePath string `json:"output_file_path"`
	Values         bool   `json:"values"`
}

func readFileData(options secretGenerateOptions) ([]byte, error) {
	if exist, err := file.FileExists(options.FilePath); err != nil {
		return nil, err
	} else if !exist {
		absFilePath, err := filepath.Abs(options.FilePath)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("secret file '%s' not found", absFilePath)
	}

	fileData, err := ioutil.ReadFile(options.FilePath)
	if err != nil {
		return nil, err
	}

	return fileData, err
}

func generateFromStdin(s *deploy.SecretGenerator) ([]byte, error) {
	var data []byte
	var secretData []byte

	var err error
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Printf("Enter secret: ")
		secretData, err = gopass.GetPasswd()
		if err != nil {
			if err.Error() == "interrupted" {
				return nil, nil
			}
			return nil, err
		}
	} else {
		r := bufio.NewReader(os.Stdin)
		buf := make([]byte, 0, 4<<20)
		for {
			n, err := r.Read(buf[:cap(buf)])
			buf = buf[:n]

			secretData = append(secretData, buf...)

			if n == 0 {
				if err == nil {
					continue
				}

				if err == io.EOF {
					break
				}

				return nil, err
			}

			if err != nil && err != io.EOF {
				return nil, err
			}
		}

		secretData = bytes.TrimSpace(secretData)
	}

	if len(secretData) == 0 {
		return nil, nil
	}

	data, err = s.Generate(secretData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func saveGeneratedData(data []byte, options secretGenerateOptions) error {
	if options.OutputFilePath != "" {
		if err := os.MkdirAll(filepath.Dir(options.OutputFilePath), 0777); err != nil {
			return err
		}

		if err := ioutil.WriteFile(options.OutputFilePath, data, 0644); err != nil {
			return err
		}
	} else {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			if !bytes.HasSuffix(data, []byte("\n")) {
				data = append(data, []byte("\n")...)
			}
		}

		fmt.Printf(string(data))
	}

	return nil
}
