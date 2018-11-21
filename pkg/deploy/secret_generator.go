package deploy

import (
	"fmt"
	"gopkg.in/yaml.v2"

	"github.com/flant/dapp/pkg/secret"
)

type SecretGenerator struct {
	generateFunc func([]byte) ([]byte, error)
}

func NewSecretEncodeGenerator(s secret.Secret) (*SecretGenerator, error) {
	g := &SecretGenerator{}

	if s != nil {
		g.generateFunc = s.Generate
	} else {
		g.generateFunc = doNothing
	}

	return g, nil
}

func NewSecretDecodeGenerator(s secret.Secret) (*SecretGenerator, error) {
	generator := &SecretGenerator{}

	if s != nil {
		generator.generateFunc = s.Extract
	} else {
		generator.generateFunc = doNothing
	}

	return generator, nil
}

func doNothing(_ []byte) ([]byte, error) { return []byte{}, nil }

func (s *SecretGenerator) Generate(data []byte) ([]byte, error) {
	return s.generateFunc(data)
}

func (s *SecretGenerator) GenerateYamlData(data []byte) ([]byte, error) {
	config := make(yaml.MapSlice, 0)
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	resultConfig, err := s.generateYamlValue(config)
	if err != nil {
		return nil, err
	}

	resultData, err := yaml.Marshal(resultConfig)
	if err != nil {
		return nil, err
	}

	return resultData, nil
}

func (s *SecretGenerator) generateYamlValue(data interface{}) (interface{}, error) {
	switch data.(type) {
	case yaml.MapSlice:
		result := make(yaml.MapSlice, len(data.(yaml.MapSlice)))
		for ind, elm := range data.(yaml.MapSlice) {
			result[ind].Key = elm.Key
			resultValue, err := s.generateYamlValue(elm.Value)
			if err != nil {
				return nil, err
			}

			result[ind].Value = resultValue
		}

		return result, nil
	case yaml.MapItem:
		var result yaml.MapItem

		result.Key = data.(yaml.MapItem).Key

		resultValue, err := s.generateYamlValue(data.(yaml.MapItem).Value)
		if err != nil {
			return nil, err
		}

		result.Value = resultValue

		return result, nil
	case []interface{}:
		var result []interface{}
		for _, elm := range data.([]interface{}) {
			resultElm, err := s.generateYamlValue(elm)
			if err != nil {
				return nil, err
			}

			result = append(result, resultElm)
		}

		return result, nil
	default:
		result, err := s.Generate([]byte(fmt.Sprintf("%v", data)))
		if err != nil {
			return nil, err
		}

		return string(result), nil
	}
}
