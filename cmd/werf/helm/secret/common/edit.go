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

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

func SecretEdit(m secret.Manager, filePath string, values bool) error {
	data, encodedData, err := readEditedFile(m, filePath, values)
	if err != nil {
		return err
	}

	tmpFilePath := filepath.Join(werf.GetTmpDir(), fmt.Sprintf("werf-edit-secret-%s.yaml", uuid.NewV4().String()))
	defer os.RemoveAll(tmpFilePath)

	if err := createTmpEditedFile(tmpFilePath, data); err != nil {
		return err
	}

	bin, binArgs, err := editor()
	if err != nil {
		return err
	}

	args := append(binArgs, tmpFilePath)
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
			newEncodedData, err = m.EncryptYamlData(newData)
			if err != nil {
				return err
			}
		} else {
			newEncodedData, err = m.Encrypt(newData)
			if err != nil {
				return err
			}

			newEncodedData = append(newEncodedData, []byte("\n")...)
		}

		if !bytes.Equal(data, newData) {
			if values {
				newEncodedData, err = prepareResultValuesData(data, encodedData, newData, newEncodedData)
				if err != nil {
					return err
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
				logboek.LogErrorF("Error: %s\n", err)
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
			data, err = m.DecryptYamlData(encodedData)
			if err != nil {
				return nil, nil, err
			}
		} else {
			data, err = m.Decrypt(encodedData)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return data, encodedData, nil
}

func askForConfirmation() (bool, error) {
	r := os.Stdin

	logboek.LogHighlightLn("Do you want to continue editing the file (Y/n)?")

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

	for _, bin := range []string{"vim", "vi", "nano"} {
		cmd := exec.Command("which", bin)
		cmd.Env = os.Environ()
		if err := cmd.Run(); err != nil {
			continue
		}

		return bin, editorArgs, nil
	}

	return "", editorArgs, fmt.Errorf("editor not detected")
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
	err := yaml.UnmarshalStrict(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func mergeYamlEncodedData(d, eD, newD, newED interface{}) (interface{}, error) {
	dType := reflect.TypeOf(d)
	newDType := reflect.TypeOf(newD)

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
