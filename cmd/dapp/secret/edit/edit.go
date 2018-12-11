package secret

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/kubernetes/pkg/util/file"

	"github.com/flant/dapp/cmd/dapp/common"
	secret_common "github.com/flant/dapp/cmd/dapp/secret/common"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/deploy/secret"
)

var CmdData struct {
	Values bool
}

var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit FILE_PATH",
		Short: "Edit or create new secret file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runSecretEdit(args[0])
			if err != nil {
				return fmt.Errorf("secret edit failed: %s", err)
			}
			return nil
		},
	}

	common.SetupDir(&CommonCmdData, cmd)
	common.SetupTmpDir(&CommonCmdData, cmd)
	common.SetupHomeDir(&CommonCmdData, cmd)

	cmd.PersistentFlags().BoolVarP(&CmdData.Values, "values", "", false, "Edit FILE_PATH as secret values file")

	return cmd
}

func runSecretEdit(filepPath string) error {
	if err := dapp.Init(*CommonCmdData.TmpDir, *CommonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&CommonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	m, err := secret.GetManager(projectDir)
	if err != nil {
		return err
	}

	return secretEdit(m, filepPath, CmdData.Values)
}

func secretEdit(m secret.Manager, filePath string, values bool) error {
	data, encodedData, err := readEditedFile(m, filePath, values)
	if err != nil {
		return err
	}

	tmpFileName := "tmp_secret_file"
	if values {
		tmpFileName = "tmp_secret_file.yaml"
	}

	tmpDir, err := common.GetTmpDir()
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}

	tmpFilePath := filepath.Join(tmpDir, tmpFileName)
	if err := createTmpEditedFile(tmpFilePath, data); err != nil {
		return err
	}

	bin, err := editor()
	if err != nil {
		return err
	}

	editIteration := func() error {
		cmd := exec.Command(bin, tmpFilePath)
		cmd.Env = os.Environ()
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
			newEncodedData, err = m.GenerateYamlData(newData)
			if err != nil {
				return err
			}
		} else {
			newEncodedData, err = m.Generate(newData)
			if err != nil {
				return err
			}

			newEncodedData = append(newEncodedData, []byte("\n")...)
		}

		if !bytes.Equal(data, newData) {
			if values {
				newEncodedData, err = prepareResultValuesData(data, encodedData, newData, newEncodedData)
			}

			if err := secret_common.SaveGeneratedData(filePath, newEncodedData); err != nil {
				return err
			}
		}

		return nil
	}

	for {
		err := editIteration()
		if err != nil {
			if strings.HasPrefix(err.Error(), "encoding failed") {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
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

func readEditedFile(m secret.Manager, filePath string, values bool) ([]byte, []byte, error) {
	var data, encodedData []byte

	exist, err := file.FileExists(filePath)
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
			data, err = m.ExtractYamlData(encodedData)
			if err != nil {
				return nil, nil, err
			}
		} else {
			data, err = m.Extract(encodedData)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return data, encodedData, nil
}

func askForConfirmation() (bool, error) {
	r := os.Stdin

	fmt.Println("Do you want to continue editing the file (Y/n)?")

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
	if err := secret_common.SaveGeneratedData(filePath, data); err != nil {
		return err
	}
	return nil
}

func editor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor, nil
	}

	for _, bin := range []string{"vim", "vi", "nano"} {
		cmd := exec.Command("which", bin)
		cmd.Env = os.Environ()
		if err := cmd.Run(); err != nil {
			continue
		}

		return bin, nil
	}

	return "", fmt.Errorf("editor not detected")
}

func prepareResultValuesData(data, encodedData, newData, newEncodedData []byte) ([]byte, error) {
	dataConfig, err := unmarshalYaml(data)
	if err != nil {
		return nil, err
	}

	encodeDataConfig, err := unmarshalYaml(encodedData)
	if err != nil {
		return nil, err
	}

	newDataConfig, err := unmarshalYaml(newData)
	if err != nil {
		return nil, err
	}

	newEncodedDataConfig, err := unmarshalYaml(newEncodedData)
	if err != nil {
		return nil, err
	}

	resultEncodedDataConfig, err := mergeYamlEncodedData(dataConfig, encodeDataConfig, newDataConfig, newEncodedDataConfig)
	if err != nil {
		return nil, err
	}

	resultEncodedData, err := yaml.Marshal(&resultEncodedDataConfig)
	if err != nil {
		return nil, err
	}

	return resultEncodedData, nil
}

func unmarshalYaml(data []byte) (yaml.MapSlice, error) {
	config := make(yaml.MapSlice, 0)
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func mergeYamlEncodedData(d, eD, newD, newED interface{}) (interface{}, error) {
	dType := reflect.TypeOf(d)
	newDType := reflect.TypeOf(d)

	if dType != newDType {
		return newED, nil
	}

	switch newD.(type) {
	case yaml.MapSlice:
		newDMapSlice := newD.(yaml.MapSlice)
		dMapSlice := d.(yaml.MapSlice)
		resultMapSlice := make(yaml.MapSlice, len(newDMapSlice))

		findDMapItemByKey := func(key interface{}) (int, *yaml.MapItem) {
			for ind, elm := range dMapSlice {
				if elm.Key == key {
					return ind, &elm
				}
			}

			return 0, nil
		}

		for ind, elm := range newDMapSlice {
			newEDMapItem := newED.(yaml.MapSlice)[ind]
			resultMapItem := newEDMapItem

			dInd, dElm := findDMapItemByKey(elm.Key)
			if dElm != nil {
				eDMapItem := eD.(yaml.MapSlice)[dInd]
				result, err := mergeYamlEncodedData(dMapSlice[dInd], eDMapItem, newDMapSlice[ind], newEDMapItem)
				if err != nil {
					return nil, err
				}

				resultMapItem = result.(yaml.MapItem)
			}

			resultMapSlice[ind] = resultMapItem
		}

		return resultMapSlice, nil
	case yaml.MapItem:
		var resultMapItem yaml.MapItem
		newDMapItem := newD.(yaml.MapItem)
		newEDMapItem := newED.(yaml.MapItem)
		dMapItem := d.(yaml.MapItem)
		eDMapItem := eD.(yaml.MapItem)

		resultMapItem.Key = newDMapItem.Key

		resultValue, err := mergeYamlEncodedData(dMapItem.Value, eDMapItem.Value, newDMapItem.Value, newEDMapItem.Value)
		if err != nil {
			return nil, err
		}

		resultMapItem.Value = resultValue

		return resultMapItem, nil
	default:
		if !reflect.DeepEqual(d, newD) {
			return newED, nil
		} else {
			return eD, nil
		}
	}
}
