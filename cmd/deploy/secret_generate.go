package main

import (
	"bytes"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/secret"
)

func newSecretGenerateGenerator(s secret.Secret) (*deploy.SecretGenerator, error) {
	g, err := deploy.NewSecretEncodeGenerator(s)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func secretGenerate(s *deploy.SecretGenerator, options secretGenerateOptions) error {
	var data []byte
	var err error

	if options.FilePath != "" {
		fileData, err := readFileData(options)
		if err != nil {
			return err
		}

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
	} else {
		data, err = generateFromStdin(s)
		if err != nil {
			return err
		}

		if data == nil {
			return nil
		}
	}

	data = append(bytes.TrimSpace(data), []byte("\n")...)

	if err := saveGeneratedData(data, options); err != nil {
		return err
	}

	return nil
}
