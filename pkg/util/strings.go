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

func ExcludeFromStringArray(list []string, elmsToRemove ...string) []string {
	var resultList []string

outerLoop:
	for _, elm := range list {
		for _, elmToExclude := range elmsToRemove {
			if elmToExclude == elm {
				continue outerLoop
			}
		}

		resultList = append(resultList, elm)
	}

	return resultList
}

func AddNewStringsToStringArray(list []string, elmsToAdd ...string) []string {
outerLoop:
	for _, elmToAdd := range elmsToAdd {
		for _, elm := range list {
			if elm == elmToAdd {
				continue outerLoop
			}
		}

		list = append(list, elmToAdd)
	}

	return list
}

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
