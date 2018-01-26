package config

import (
	"fmt"
	"reflect"
	"strings"
)

func CheckOverflow(m map[string]interface{}, config interface{}) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}

		val := reflect.Indirect(reflect.ValueOf(config))
		return fmt.Errorf("в конфиге `%s` содержатся не поддерживаемые поля: `%s`", val.Type().Name(), strings.Join(keys, "`, `")) // FIXME
	}
	return nil
}

func InterfaceToStringArray(stringOrStringArray interface{}) ([]string, error) {
	if stringOrStringArray == nil {
		return []string{}, nil
	} else if val, ok := stringOrStringArray.(string); ok {
		return []string{val}, nil
	} else if interfaceArray, ok := stringOrStringArray.([]interface{}); ok {
		stringArray := []string{}
		for _, interf := range interfaceArray {
			if val, ok := interf.(string); ok {
				stringArray = append(stringArray, val)
			} else {
				return nil, fmt.Errorf("ожидается строка или массив строк: %v", stringOrStringArray) // FIXME
			}
		}
		return stringArray, nil
	} else {
		return nil, fmt.Errorf("ожидается строка или массив строк: %v", stringOrStringArray) // FIXME
	}
}
