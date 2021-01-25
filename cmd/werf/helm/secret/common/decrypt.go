package secret

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/secret"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/werf/pkg/deploy/secrets_manager"
)

func SecretFileDecrypt(ctx context.Context, m *secrets_manager.SecretsManager, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         false,
	}

	return secretDecrypt(ctx, m, options)
}

func SecretValuesDecrypt(ctx context.Context, m *secrets_manager.SecretsManager, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         true,
	}

	return secretDecrypt(ctx, m, options)
}

func secretDecrypt(ctx context.Context, m *secrets_manager.SecretsManager, options *GenerateOptions) error {
	var encodedData []byte
	var data []byte
	var err error

	var encoder *secret.YamlEncoder
	if enc, err := m.GetYamlEncoder(ctx); err != nil {
		return err
	} else {
		encoder = enc
	}

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
		data, err = encoder.DecryptYamlData(encodedData)
		if err != nil {
			return err
		}
	} else {
		data, err = encoder.Decrypt(encodedData)
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
