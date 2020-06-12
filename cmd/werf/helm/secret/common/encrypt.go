package secret

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/werf/pkg/deploy/secret"
)

func SecretFileEncrypt(m secret.Manager, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         false,
	}

	return secretEncrypt(m, options)
}

func SecretValuesEncrypt(m secret.Manager, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         true,
	}

	return secretEncrypt(m, options)
}

func secretEncrypt(m secret.Manager, options *GenerateOptions) error {
	var data []byte
	var encodedData []byte
	var err error

	if options.FilePath != "" {
		data, err = ReadFileData(options.FilePath)
		if err != nil {
			return err
		}
	} else if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		data, err = InputFromStdin()
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return nil
		}
	} else {
		return ExpectedFilePathOrPipeError()
	}

	if options.Values {
		encodedData, err = m.EncryptYamlData(data)
		if err != nil {
			return err
		}
	} else {
		encodedData, err = m.Encrypt(data)
		if err != nil {
			return err
		}
	}

	if !bytes.HasSuffix(encodedData, []byte("\n")) {
		encodedData = append(encodedData, []byte("\n")...)
	}

	if options.OutputFilePath != "" {
		if err := SaveGeneratedData(options.OutputFilePath, encodedData); err != nil {
			return err
		}
	} else {
		fmt.Printf("%s", string(encodedData))
	}

	return nil
}
