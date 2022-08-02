package secret

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/secret"
)

func SecretFileEncrypt(ctx context.Context, m *secrets_manager.SecretsManager, workingDir, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         false,
	}

	return secretEncrypt(ctx, m, workingDir, options)
}

func SecretValuesEncrypt(ctx context.Context, m *secrets_manager.SecretsManager, workingDir, filePath, outputFilePath string) error {
	options := &GenerateOptions{
		FilePath:       filePath,
		OutputFilePath: outputFilePath,
		Values:         true,
	}

	return secretEncrypt(ctx, m, workingDir, options)
}

func secretEncrypt(ctx context.Context, m *secrets_manager.SecretsManager, workingDir string, options *GenerateOptions) error {
	var data []byte
	var encodedData []byte
	var err error

	var encoder *secret.YamlEncoder
	if enc, err := m.GetYamlEncoder(ctx, workingDir); err != nil {
		return err
	} else {
		encoder = enc
	}

	switch {
	case options.FilePath != "":
		data, err = ReadFileData(options.FilePath)
		if err != nil {
			return err
		}
	case !terminal.IsTerminal(int(os.Stdin.Fd())):
		data, err = InputFromStdin()
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return nil
		}
	default:
		return ExpectedFilePathOrPipeError()
	}

	if options.Values {
		encodedData, err = encoder.EncryptYamlData(data)
		if err != nil {
			return err
		}
	} else {
		encodedData, err = encoder.Encrypt(data)
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
