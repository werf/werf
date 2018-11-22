package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/secret"

	"golang.org/x/crypto/ssh/terminal"
)

func newSecretExtractGenerator(s secret.Secret) (*deploy.SecretGenerator, error) {
	g, err := deploy.NewSecretDecodeGenerator(s)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func secretExtract(s *deploy.SecretGenerator, options secretGenerateOptions) error {
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
		data, err = s.GenerateYamlData(encodedData)
		if err != nil {
			return fmt.Errorf("check encryption key and data: %s", err)
		}
	} else {
		data, err = s.Generate(encodedData)
		if err != nil {
			return fmt.Errorf("check encryption key and data: %s", err)
		}
	}

	if options.OutputFilePath != "" {
		if err := saveGeneratedData(options.OutputFilePath, data, options); err != nil {
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
