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

func MapValues[M ~map[K]V, K comparable, V any](m M) (res []V) {
	for _, v := range m {
		res = append(res, v)
	}
	return
}

func MapKeys[M ~map[K]V, K comparable, V any](m M) (res []K) {
	for k := range m {
		res = append(res, k)
	}
	return
}

func SortedStringKeys[T map[string]any](m T) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
