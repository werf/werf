package util

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
