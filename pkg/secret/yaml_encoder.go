package secret

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// YamlEncoder is an Encoder compatible object with additional helpers to work with yaml data: EncryptYamlData and DecryptYamlData
type YamlEncoder struct {
	Encoder Encoder

	generateFunc func([]byte) ([]byte, error)
	extractFunc  func([]byte) ([]byte, error)
}

func NewYamlEncoder(encoder Encoder) *YamlEncoder {
	yamlEncoder := &YamlEncoder{Encoder: encoder}

	if encoder != nil {
		yamlEncoder.generateFunc = encoder.Encrypt
		yamlEncoder.extractFunc = encoder.Decrypt
	} else {
		yamlEncoder.generateFunc = doNothing
		yamlEncoder.extractFunc = doNothing
	}

	return yamlEncoder
}

func (s *YamlEncoder) Encrypt(data []byte) ([]byte, error) {
	resultData, err := s.generateFunc(data)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func (s *YamlEncoder) EncryptYamlData(data []byte) ([]byte, error) {
	resultData, err := doYamlData(s.generateFunc, data)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func (s *YamlEncoder) Decrypt(data []byte) ([]byte, error) {
	resultData, err := s.extractFunc(data)
	if err != nil {
		if IsExtractDataError(err) {
			return nil, fmt.Errorf("decryption failed: check data `%s`: %s", string(data), err)
		}

		return nil, fmt.Errorf("decryption failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func (s *YamlEncoder) DecryptYamlData(data []byte) ([]byte, error) {
	resultData, err := doYamlData(s.extractFunc, data)
	if err != nil {
		if IsExtractDataError(err) {
			return nil, fmt.Errorf("decryption failed: check data `%s`: %s", string(data), err)
		}

		return nil, fmt.Errorf("decryption failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func doYamlData(doFunc func([]byte) ([]byte, error), data []byte) ([]byte, error) {
	config := make(yaml.MapSlice, 0)
	err := yaml.UnmarshalStrict(data, &config)
	if err != nil {
		return nil, err
	}

	resultConfig, err := doYamlValueSecret(doFunc, config)
	if err != nil {
		return nil, err
	}

	resultData, err := yaml.Marshal(resultConfig)
	if err != nil {
		return nil, err
	}

	return resultData, nil
}

func doYamlValueSecret(doFunc func([]byte) ([]byte, error), data interface{}) (interface{}, error) {
	switch data.(type) {
	case yaml.MapSlice:
		result := make(yaml.MapSlice, len(data.(yaml.MapSlice)))
		for ind, elm := range data.(yaml.MapSlice) {
			result[ind].Key = elm.Key
			resultValue, err := doYamlValueSecret(doFunc, elm.Value)
			if err != nil {
				return nil, err
			}

			result[ind].Value = resultValue
		}

		return result, nil
	case yaml.MapItem:
		var result yaml.MapItem

		result.Key = data.(yaml.MapItem).Key

		resultValue, err := doYamlValueSecret(doFunc, data.(yaml.MapItem).Value)
		if err != nil {
			return nil, err
		}

		result.Value = resultValue

		return result, nil
	case []interface{}:
		var result []interface{}
		for _, elm := range data.([]interface{}) {
			resultElm, err := doYamlValueSecret(doFunc, elm)
			if err != nil {
				return nil, err
			}

			result = append(result, resultElm)
		}

		return result, nil
	default:
		result, err := doFunc([]byte(fmt.Sprintf("%v", data)))
		if err != nil {
			return nil, err
		}

		return string(result), nil
	}
}

func doNothing(data []byte) ([]byte, error) { return data, nil }
