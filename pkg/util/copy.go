package util

func CopyArr[T any](arr []T) (ret []T) {
	ret = append(ret, arr...)
	return
}

func CopyMap[K comparable, V any](m map[K]V) (ret map[K]V) {
	ret = make(map[K]V)
	for k, v := range m {
		ret[k] = v
	}
	return
}
