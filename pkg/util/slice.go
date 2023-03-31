package util

func MapFuncToSlice[T, RT any, FT func(T) RT](arr []T, f FT) (res []RT) {
	for _, v := range arr {
		res = append(res, f(v))
	}
	return
}
