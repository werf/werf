package util

func MapFuncToSlice[T, RT any, FT func(T) RT](arr []T, f FT) (res []RT) {
	for _, v := range arr {
		res = append(res, f(v))
	}
	return
}

func SliceToMapWithValue[K comparable, V any](keys []K, value V) map[K]V {
	resultMap := make(map[K]V)
	for _, key := range keys {
		resultMap[key] = value
	}
	return resultMap
}
