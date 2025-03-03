package common

import (
	"fmt"
	"testing"
)

func TestKeyValueArrayToMap_Positive(t *testing.T) {
	tests := []struct {
		input    []string
		expected map[string]string
	}{
		{
			input: []string{"key1=value1"},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			input: []string{"key1=value1", "key2=value2"},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			input: []string{"key1=value1,key2=value2"},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			input: []string{"key1=value1,key2="},
			expected: map[string]string{
				"key1": "value1",
				"key2": "",
			},
		},
		{
			input: []string{"key=value with pair separator="},
			expected: map[string]string{
				"key": "value with pair separator=",
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("input=%v", test.input), func(t *testing.T) {
			result, err := InputArrayToKeyValueMap(test.input, DefaultPairSeparator, DefaultKeyValueSeparator)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !mapsEqual(result, test.expected) {
				t.Errorf("expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestKeyValueArrayToMap_Negative(t *testing.T) {
	inputs := [][]string{
		{""},
		{"=value"},
		{","},
		{"key=value,with value separator"},
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("input=%v", input), func(t *testing.T) {
			_, err := InputArrayToKeyValueMap(input, DefaultPairSeparator, DefaultKeyValueSeparator)
			if err == nil {
				t.Errorf("expected error but got nil")
			}
		})
	}
}

// mapsEqual compares two maps and checks for exact equality.
func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}
