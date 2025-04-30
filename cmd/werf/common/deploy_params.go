package common

import (
	"fmt"
	"strings"

	"github.com/werf/common-go/pkg/util"
)

const (
	DefaultKeyValueSeparator = "="
	DefaultPairSeparator     = ","
)

func GetUserExtraAnnotations(cmdData *CmdData) (map[string]string, error) {
	result, err := InputArrayToKeyValueMap(GetAddAnnotations(cmdData), *cmdData.AddAnnotationSeparator, DefaultKeyValueSeparator)
	if err != nil {
		return nil, fmt.Errorf("unsupported --add-annotation value: %w", err)
	}

	return result, nil
}

func GetUserExtraLabels(cmdData *CmdData) (map[string]string, error) {
	result, err := InputArrayToKeyValueMap(GetAddLabels(cmdData), *cmdData.AddLabelSeparator, DefaultKeyValueSeparator)
	if err != nil {
		return nil, fmt.Errorf("unsupported --add-label value: %w", err)
	}

	return result, nil
}

// InputArrayToKeyValueMap converts an array of strings in the form of key1=value1[,key2=value2] to a map.
func InputArrayToKeyValueMap(input []string, pairSep, keyValueSep string) (map[string]string, error) {
	result := map[string]string{}
	for _, value := range input {
		pairs := strings.Split(value, pairSep)
		valueResult, err := keyValueArrayToMap(pairs, keyValueSep)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q (%q): %w", value, pairSep, err)
		}

		result = util.MergeMaps(result, valueResult)
	}

	return result, nil
}

func keyValueArrayToMap(pairs []string, sep string) (map[string]string, error) {
	result := map[string]string{}
	for _, pair := range pairs {
		parts := strings.SplitN(pair, sep, 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid key=value pair %q (%q)", pair, sep)
		}

		result[parts[0]] = parts[1]
	}

	return result, nil
}
