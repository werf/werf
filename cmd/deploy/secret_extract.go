package main

import (
	"bytes"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/secret"
)

func newSecretExtractGenerator(s secret.Secret) (*deploy.SecretGenerator, error) {
	g, err := deploy.NewSecretDecodeGenerator(s)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func secretExtract(s *deploy.SecretGenerator, options secretGenerateOptions) error {
	var data []byte
	var err error

	if options.FilePath != "" {
		fileData, err := readFileData(options)
		if err != nil {
			return err
		}

		fileData = bytes.TrimSpace(fileData)

		if options.Values {
			data, err = s.GenerateYamlData(fileData)
			if err != nil {
				return err
			}
		} else {
			data, err = s.Generate(fileData)
			if err != nil {
				return err
			}
		}

		if err := saveGeneratedData(data, options); err != nil {
			return err
		}
	} else {
		data, err = generateFromStdin(s)
		if err != nil {
			return err
		}

		if data == nil {
			return nil
		}

		if err := saveGeneratedData(data, options); err != nil {
			return err
		}
	}

	return nil
}
