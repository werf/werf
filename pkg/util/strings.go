package util

import "sort"

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

func FilterSlice[V any](slice []V, filterFunc func(i int, val V) bool) []V {
	var result []V
	for i, val := range slice {
		if filterFunc(i, val) {
			result = append(result, val)
		}
	}

	return result
}

// Returns nil if no match.
func FirstMatchInSliceIndex[V any](slice []V, matchFunc func(i int, val V) bool) *int {
	for i := 0; i < len(slice); i++ {
		if matchFunc(i, slice[i]) {
			return &i
		}
	}

	return nil
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

func FindDuplicatedStrings(elems []string) []string {
	if len(elems) <= 1 {
		return []string{}
	}

	sort.Strings(elems)

	var duplicates []string
	for i, elem := range elems {
		if i == 0 {
			continue
		}

		if elem == elems[i-1] {
			duplicates = append(duplicates, elem)
		}
	}

	return duplicates
}
