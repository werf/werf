package secret

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type BaseSecret struct {
	generateFunc func([]byte) ([]byte, error)
	extractFunc  func([]byte) ([]byte, error)
}

func doNothing(_ []byte) ([]byte, error) { return []byte{}, nil }

func (s *BaseSecret) Generate(data []byte) ([]byte, error) {
	resultData, err := s.generateFunc(data)
	if err != nil {
		return nil, fmt.Errorf("encoding failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func (s *BaseSecret) GenerateYamlData(data []byte) ([]byte, error) {
	resultData, err := doYamlData(s.generateFunc, data)
	if err != nil {
		return nil, fmt.Errorf("encoding failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func (s *BaseSecret) Extract(data []byte) ([]byte, error) {
	resultData, err := s.extractFunc(data)
	if err != nil {
		return nil, fmt.Errorf("decoding failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func (s *BaseSecret) ExtractYamlData(data []byte) ([]byte, error) {
	resultData, err := doYamlData(s.extractFunc, data)
	if err != nil {
		return nil, fmt.Errorf("decoding failed: check encryption key and data: %s", err)
	}

	return resultData, nil
}

func doYamlData(doFunc func([]byte) ([]byte, error), data []byte) ([]byte, error) {
	config := make(yaml.MapSlice, 0)
	err := yaml.Unmarshal(data, &config)
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
