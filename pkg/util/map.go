package util

import (
	"sort"
)

// Dest has higher priority.
func MergeMaps[K comparable, V any](src, dest map[K]V) map[K]V {
	result := make(map[K]V)

	for k, v := range src {
		result[k] = v
	}

	for k, v := range dest {
		result[k] = v
	}

	return result
}

func SortedStringKeys(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
