package util

import (
	"fmt"
)

func MapStringInterfaceToMapStringString(value map[string]interface{}) map[string]string {
	result := map[string]string{}
	for key, valueInterf := range value {
		result[key] = fmt.Sprintf("%v", valueInterf)
	}

	return result
}

func InterfaceToStringArray(value interface{}) ([]string, error) {
	switch value := value.(type) {
	case []interface{}:
		result, err := InterfaceArrayToStringArray(value)
		if err != nil {
			return nil, err
		}
		return result, nil
	case []string:
		return value, nil
	default:
		return nil, fmt.Errorf("value `%#v` can't be casted into []string", value)
	}
}

func InterfaceArrayToStringArray(array []interface{}) ([]string, error) {
	var result []string
	for _, value := range array {
		if str, ok := value.(string); !ok {
			return nil, fmt.Errorf("value `%#v` can't be casted into string", value)
		} else {
			result = append(result, str)
		}
	}
	return result, nil
}

func InterfaceToMapStringInterface(value interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	switch value := value.(type) {
	case map[string]interface{}:
		return value, nil
	case map[interface{}]interface{}:
		for k, v := range value {
			key, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("key value `%#v` can't be casted into string", key)
			}
			result[key] = v
		}
		return result, nil
	default:
		return nil, fmt.Errorf("value `%#v` can't be casted into map[string]interface{}", value)
	}
}
