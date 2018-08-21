package util

import "fmt"

func InterfaceToStringArray(value interface{}) ([]string, error) {
	switch value.(type) {
	case []interface{}:
		result, err := InterfaceArrayToStringArray(value.([]interface{}))
		if err != nil {
			return nil, err
		}
		return result, nil
	case []string:
		return value.([]string), nil
	default:
		return nil, fmt.Errorf("value `%#v` can't be casting into []string", value)
	}
}

func InterfaceArrayToStringArray(array []interface{}) ([]string, error) {
	var result []string
	for _, value := range array {
		if str, ok := value.(string); !ok {
			return nil, fmt.Errorf("value `%#v` can't be casting into string", value)
		} else {
			result = append(result, str)
		}
	}
	return result, nil
}

func InterfaceToMapStringInterface(value interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	switch value.(type) {
	case map[string]interface{}:
		return value.(map[string]interface{}), nil
	case map[interface{}]interface{}:
		for k, v := range value.(map[interface{}]interface{}) {
			key, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("key value `%#v` can't be casting into string", key)
			}
			result[key] = v
		}
		return result, nil
	default:
		return nil, fmt.Errorf("value `%#v` can't be casting into map[string]interface{}", value)
	}
}
