package secret

import (
	"bytes"
	"fmt"

	yaml_v3 "gopkg.in/yaml.v3"
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
		return nil, fmt.Errorf("encryption failed: check encryption key and data: %w", err)
	}

	return resultData, nil
}

func (s *YamlEncoder) EncryptYamlData(data []byte) ([]byte, error) {
	resultData, err := doYamlDataV2(s.generateFunc, data)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: check encryption key and data: %w", err)
	}

	return resultData, nil
}

func (s *YamlEncoder) Decrypt(data []byte) ([]byte, error) {
	resultData, err := s.extractFunc(data)
	if err != nil {
		if IsExtractDataError(err) {
			return nil, fmt.Errorf("decryption failed: check data `%s`: %w", string(data), err)
		}

		return nil, fmt.Errorf("decryption failed: check encryption key and data: %w", err)
	}

	return resultData, nil
}

func (s *YamlEncoder) DecryptYamlData(data []byte) ([]byte, error) {
	resultData, err := doYamlDataV2(s.extractFunc, data)
	if err != nil {
		if IsExtractDataError(err) {
			return nil, fmt.Errorf("decryption failed: check data `%s`: %w", string(data), err)
		}

		return nil, fmt.Errorf("decryption failed: check encryption key and data: %w", err)
	}

	return resultData, nil
}

func doYamlDataV2(doFunc func([]byte) ([]byte, error), data []byte) ([]byte, error) {
	var config yaml_v3.Node

	if err := yaml_v3.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config data: %w", err)
	}

	resultConfig, err := doYamlValueSecretV2(doFunc, &config)
	if err != nil {
		return nil, fmt.Errorf("unable to process config secrets: %w", err)
	}

	var resultData bytes.Buffer

	yamlEncoder := yaml_v3.NewEncoder(&resultData)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(resultConfig); err != nil {
		return nil, fmt.Errorf("unable to marshal modified config data: %w", err)
	}

	return resultData.Bytes(), nil
}

func doYamlValueSecretV2(doFunc func([]byte) ([]byte, error), node *yaml_v3.Node) (*yaml_v3.Node, error) {
	switch node.Kind {
	case yaml_v3.DocumentNode:
		for pos := 0; pos < len(node.Content); pos += 1 {
			newValueNode, err := doYamlValueSecretV2(doFunc, node.Content[pos])
			if err != nil {
				return nil, fmt.Errorf("unable to process document key %d: %w", pos, err)
			}
			node.Content[pos] = newValueNode
		}

	case yaml_v3.MappingNode:
		for pos := 0; pos < len(node.Content); pos += 2 {
			keyNode := node.Content[pos]
			valueNode := node.Content[pos+1]
			newValueNode, err := doYamlValueSecretV2(doFunc, valueNode)
			if err != nil {
				return nil, fmt.Errorf("unable to process map key %q: %w", keyNode.Value, err)
			}
			node.Content[pos+1] = newValueNode
		}

	case yaml_v3.SequenceNode:
		for pos := 0; pos < len(node.Content); pos += 1 {
			newValueNode, err := doYamlValueSecretV2(doFunc, node.Content[pos])
			if err != nil {
				return nil, fmt.Errorf("unable to process array key %d: %w", pos, err)
			}
			node.Content[pos] = newValueNode
		}

	case yaml_v3.AliasNode:
		newAliasNode, err := doYamlValueSecretV2(doFunc, node.Alias)
		if err != nil {
			return nil, fmt.Errorf("unable to process an alias node %q: %w", node.Value, err)
		}
		node.Alias = newAliasNode

	case yaml_v3.ScalarNode:
		if node.ShortTag() == "!!str" {
			var value string

			if err := node.Decode(&value); err != nil {
				return nil, fmt.Errorf("unable to decode string value %q: %w", node.Value, err)
			}

			newValue, err := doFunc([]byte(value))
			if err != nil {
				return nil, err
			}

			if err := node.Encode(string(newValue)); err != nil {
				return nil, fmt.Errorf("unable to encode string value %q: %w", string(newValue), err)
			}
		}
	}

	return node, nil
}

func doNothing(data []byte) ([]byte, error) { return data, nil }
