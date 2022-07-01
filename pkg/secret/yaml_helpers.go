package secret

import (
	"bytes"
	"fmt"

	yaml_v3 "gopkg.in/yaml.v3"
)

func MergeEncodedYaml(oldData, newData, oldEncodedData, newEncodedData []byte) ([]byte, error) {
	var oldConfig, newConfig, oldEncodedConfig, newEncodedConfig yaml_v3.Node

	for _, d := range []struct {
		Data []byte
		Node *yaml_v3.Node
	}{
		{Data: oldData, Node: &oldConfig},
		{Data: newData, Node: &newConfig},
		{Data: oldEncodedData, Node: &oldEncodedConfig},
		{Data: newEncodedData, Node: &newEncodedConfig},
	} {
		if err := yaml_v3.Unmarshal(d.Data, d.Node); err != nil {
			return nil, fmt.Errorf("unable to unmarshal yaml data: %w", err)
		}
	}

	mergedNode, err := MergeEncodedYamlNode(&oldConfig, &newConfig, &oldEncodedConfig, &newEncodedConfig)
	if err != nil {
		return nil, err
	}

	var resultData bytes.Buffer

	yamlEncoder := yaml_v3.NewEncoder(&resultData)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(mergedNode); err != nil {
		return nil, fmt.Errorf("unable to marshal merged encoded data: %w", err)
	}

	return resultData.Bytes(), nil
}

func MergeEncodedYamlNode(oldConfig, newConfig, oldEncodedConfig, newEncodedConfig *yaml_v3.Node) (*yaml_v3.Node, error) {
	if oldConfig == nil {
		return newEncodedConfig, nil
	}
	if newConfig.Kind != oldConfig.Kind {
		return newEncodedConfig, nil
	}

	switch newEncodedConfig.Kind {
	case yaml_v3.DocumentNode:
		for pos := 0; pos < len(newEncodedConfig.Content); pos += 1 {
			newEncodedValue := newEncodedConfig.Content[pos]
			newValue := newConfig.Content[pos]
			oldEncodedValue := getSubNodeByIndex(oldEncodedConfig, pos)
			oldValue := getSubNodeByIndex(oldConfig, pos)

			newValueNode, err := MergeEncodedYamlNode(oldValue, newValue, oldEncodedValue, newEncodedValue)
			if err != nil {
				return nil, fmt.Errorf("unable to process document key %d: %w", pos, err)
			}
			newEncodedConfig.Content[pos] = newValueNode
		}

	case yaml_v3.MappingNode:
		for pos := 0; pos < len(newEncodedConfig.Content); pos += 2 {
			newKey := newEncodedConfig.Content[pos]

			newEncodedValue := newEncodedConfig.Content[pos+1]
			newValue := newConfig.Content[pos+1]
			oldEncodedValue := getSubNodeByKey(oldEncodedConfig, newKey.Value)
			oldValue := getSubNodeByKey(oldConfig, newKey.Value)

			newValueNode, err := MergeEncodedYamlNode(oldValue, newValue, oldEncodedValue, newEncodedValue)
			if err != nil {
				return nil, fmt.Errorf("unable to process map key %q: %w", newEncodedConfig.Content[pos].Value, err)
			}
			newEncodedConfig.Content[pos+1] = newValueNode
		}

	case yaml_v3.SequenceNode:
		for pos := 0; pos < len(newEncodedConfig.Content); pos += 1 {
			newEncodedValue := newEncodedConfig.Content[pos]
			newValue := newConfig.Content[pos]
			oldEncodedValue := getSubNodeByIndex(oldEncodedConfig, pos)
			oldValue := getSubNodeByIndex(oldConfig, pos)

			newValueNode, err := MergeEncodedYamlNode(oldValue, newValue, oldEncodedValue, newEncodedValue)
			if err != nil {
				return nil, fmt.Errorf("unable to process array key %d: %w", pos, err)
			}
			newEncodedConfig.Content[pos] = newValueNode
		}

	case yaml_v3.AliasNode:
		newAliasNode, err := MergeEncodedYamlNode(oldConfig.Alias, newConfig.Alias, oldEncodedConfig.Alias, newEncodedConfig.Alias)
		if err != nil {
			return nil, fmt.Errorf("unable to process an alias node %q: %w", newEncodedConfig.Value, err)
		}
		newEncodedConfig.Alias = newAliasNode

	case yaml_v3.ScalarNode:
		if oldConfig.Value == newConfig.Value {
			return oldEncodedConfig, nil
		}
		return newEncodedConfig, nil
	}

	return newEncodedConfig, nil
}

func getSubNodeByIndex(node *yaml_v3.Node, ind int) *yaml_v3.Node {
	if ind < len(node.Content) {
		return node.Content[ind]
	}
	return nil
}

func getSubNodeByKey(node *yaml_v3.Node, rawKey string) *yaml_v3.Node {
	for i := 0; i < len(node.Content); i += 2 {
		k, v := node.Content[i], node.Content[i+1]
		if k.Value == rawKey {
			return v
		}
	}
	return nil
}
