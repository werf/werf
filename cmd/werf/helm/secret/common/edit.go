package secret

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/secret"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

func SecretEdit(ctx context.Context, m *secrets_manager.SecretsManager, workingDir, filePath string, values bool) error {
	var encoder *secret.YamlEncoder
	if enc, err := m.GetYamlEncoder(ctx, workingDir); err != nil {
		return err
	} else {
		encoder = enc
	}

	data, encodedData, err := readEditedFile(filePath, values, encoder)
	if err != nil {
		return err
	}

	tmpFilePath := filepath.Join(werf.GetTmpDir(), fmt.Sprintf("werf-edit-secret-%s.yaml", uuid.NewString()))
	defer os.RemoveAll(tmpFilePath)

	if err := createTmpEditedFile(tmpFilePath, data); err != nil {
		return err
	}

	bin, binArgs, err := editor()
	if err != nil {
		return err
	}

	args := binArgs
	args = append(args, tmpFilePath)
	editIteration := func() error {
		cmd := exec.Command(bin, args...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}

		newData, err := ioutil.ReadFile(tmpFilePath)
		if err != nil {
			return err
		}

		var newEncodedData []byte
		if values {
			newEncodedData, err = encoder.EncryptYamlData(newData)
			if err != nil {
				return err
			}
		} else {
			newEncodedData, err = encoder.Encrypt(newData)
			if err != nil {
				return err
			}

			newEncodedData = append(newEncodedData, []byte("\n")...)
		}

		if !bytes.Equal(data, newData) {
			if values {
				newEncodedData, err = secret.MergeEncodedYaml(data, newData, encodedData, newEncodedData)
				if err != nil {
					return fmt.Errorf("unable to merge changed values of encoded yaml: %w", err)
				}
			}

			if err := SaveGeneratedData(filePath, newEncodedData); err != nil {
				return err
			}
		}

		return nil
	}

	for {
		err := editIteration()
		if err != nil {
			if strings.HasPrefix(err.Error(), "encryption failed") {
				logboek.Warn().LogF("Error: %s\n", err)
				ok, err := askForConfirmation()
				if err != nil {
					return err
				}

				if ok {
					continue
				}
			}

			return err
		}

		break
	}

	return nil
}

func readEditedFile(filePath string, values bool, encoder *secret.YamlEncoder) ([]byte, []byte, error) {
	var data, encodedData []byte

	exist, err := util.FileExists(filePath)
	if err != nil {
		return nil, nil, err
	}

	if exist {
		encodedData, err = ioutil.ReadFile(filePath)
		if err != nil {
			return nil, nil, err
		}

		encodedData = bytes.TrimSpace(encodedData)

		if values {
			data, err = encoder.DecryptYamlData(encodedData)
			if err != nil {
				return nil, nil, err
			}
		} else {
			data, err = encoder.Decrypt(encodedData)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return data, encodedData, nil
}

func askForConfirmation() (bool, error) {
	r := os.Stdin

	fmt.Println(logboek.Colorize(style.Highlight(), "Do you want to continue editing the file (Y/n)?"))

	isTerminal := terminal.IsTerminal(int(r.Fd()))
	if isTerminal {
		if oldState, err := terminal.MakeRaw(int(r.Fd())); err != nil {
			return false, err
		} else {
			defer terminal.Restore(int(r.Fd()), oldState)
		}
	}

	var buf [1]byte
	n, err := r.Read(buf[:])
	if n > 0 {
		switch buf[0] {
		case 'y', 'Y', 13:
			return true, nil
		default:
			return false, nil
		}
	}

	if err != nil && err != io.EOF {
		return false, err
	}

	return false, nil
}

func createTmpEditedFile(filePath string, data []byte) error {
	if err := SaveGeneratedData(filePath, data); err != nil {
		return err
	}
	return nil
}

func editor() (string, []string, error) {
	var editorArgs []string

	editorValue := os.Getenv("EDITOR")
	if editorValue != "" {
		editorFields := strings.Fields(editorValue)
		return editorFields[0], editorFields[1:], nil
	}

	var defaultEditors []string
	if runtime.GOOS == "windows" {
		defaultEditors = []string{"notepad"}
	} else {
		defaultEditors = []string{"vim", "vi", "nano"}
	}

	for _, bin := range defaultEditors {
		if _, err := exec.LookPath(bin); err != nil {
			continue
		}

		return bin, editorArgs, nil
	}

	return "", editorArgs, fmt.Errorf("editor not detected")
}
