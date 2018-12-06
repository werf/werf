package util

func UniqStrings(arr []string) []string {
	res := []string{}

IterateAllValues:
	for _, v1 := range arr {
		for _, v2 := range res {
			if v1 == v2 {
				continue IterateAllValues
			}
		}

		res = append(res, v1)
	}

	return res
}

func UniqAppendString(arr []string, value string) []string {
	return UniqStrings(append(arr, value))
}

func RejectEmptyStrings(arr []string) []string {
	res := []string{}

	for _, v := range arr {
		if v == "" {
			continue
		}
		res = append(res, v)
	}

	return res
}

func IsStringsContainValue(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}
