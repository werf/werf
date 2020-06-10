package secret

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/werf/pkg/deploy/secret"
)

func SecretFileDecrypt(m secret.Manager, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         false,
	}

	return secretDecrypt(m, options)
}

func SecretValuesDecrypt(m secret.Manager, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         true,
	}

	return secretDecrypt(m, options)
}

func secretDecrypt(m secret.Manager, options *GenerateOptions) error {
	var encodedData []byte
	var data []byte
	var err error

	if options.FilePath != "" {
		encodedData, err = ReadFileData(options.FilePath)
		if err != nil {
			return err
		}
	} else {
		if !terminal.IsTerminal(int(os.Stdin.Fd())) {
			encodedData, err = InputFromStdin()
			if err != nil {
				return err
			}
		} else {
			return ExpectedFilePathOrPipeError()
		}

		if len(encodedData) == 0 {
			return nil
		}
	}

	encodedData = bytes.TrimSpace(encodedData)

	if options.Values {
		data, err = m.DecryptYamlData(encodedData)
		if err != nil {
			return err
		}
	} else {
		data, err = m.Decrypt(encodedData)
		if err != nil {
			return err
		}
	}

	if options.OutputFilePath != "" {
		if err := SaveGeneratedData(options.OutputFilePath, data); err != nil {
			return err
		}
	} else {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			if !bytes.HasSuffix(data, []byte("\n")) {
				data = append(data, []byte("\n")...)
			}
		}

		fmt.Printf("%s", string(data))
	}

	return nil
}

func ExpectedFilePathOrPipeError() error {
	return errors.New("expected FILE_PATH or pipe")
}
