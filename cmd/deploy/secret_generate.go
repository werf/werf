package main

import (
	"bytes"
	"fmt"

	"github.com/flant/dapp/pkg/deploy/secret"
)

func secretGenerate(m secret.Manager, options secretGenerateOptions) error {
	var data []byte
	var encodedData []byte
	var err error

	if options.FilePath != "" {
		data, err = readFileData(options.FilePath)
		if err != nil {
			return err
		}
	} else {
		data, err = readStdin()
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return nil
		}
	}

	if options.FilePath != "" && options.Values {
		encodedData, err = m.GenerateYamlData(data)
		if err != nil {
			return err
		}
	} else {
		encodedData, err = m.Generate(data)
		if err != nil {
			return err
		}
	}

	if !bytes.HasSuffix(encodedData, []byte("\n")) {
		encodedData = append(encodedData, []byte("\n")...)
	}

	if options.OutputFilePath != "" {
		if err := saveGeneratedData(options.OutputFilePath, encodedData); err != nil {
			return err
		}
	} else {
		fmt.Printf(string(encodedData))
	}

	return nil
}
