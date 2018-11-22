package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/deploy/secret"

	"golang.org/x/crypto/ssh/terminal"
)

func secretExtract(s secret.Secret, options secretGenerateOptions) error {
	var encodedData []byte
	var data []byte
	var err error

	if options.FilePath != "" {
		encodedData, err = readFileData(options.FilePath)
		if err != nil {
			return err
		}
	} else {
		encodedData, err = readStdin()
		if err != nil {
			return err
		}

		if len(encodedData) == 0 {
			return nil
		}
	}

	encodedData = bytes.TrimSpace(encodedData)

	if options.FilePath != "" && options.Values {
		data, err = s.ExtractYamlData(encodedData)
		if err != nil {
			return err
		}
	} else {
		data, err = s.Extract(encodedData)
		if err != nil {
			return err
		}
	}

	if options.OutputFilePath != "" {
		if err := saveGeneratedData(options.OutputFilePath, data); err != nil {
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
