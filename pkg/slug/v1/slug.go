package v1

import (
	"strings"

	"github.com/flant/dapp/pkg/util"
)

const SlugSeparator = "-"

func Slug(data string) string {
	if ShouldNotBeSlugged(data) {
		return data
	}

	sluggedData := Slugify(data)
	murmurHash := util.MurmurHash(data)

	var slugParts []string
	if sluggedData != "" {
		slugParts = append(slugParts, sluggedData)
	}
	slugParts = append(slugParts, murmurHash)

	consistentUniqSlug := strings.Join(slugParts, SlugSeparator)

	return consistentUniqSlug
}

func ShouldNotBeSlugged(data string) bool {
	return Slugify(data) == data && data != ""
}

func Slugify(data string) string {
	var result []rune

	var isCursorDash bool
	var isPreviousDash bool
	var isStartedDash, isDoubledDash bool

	isResultEmpty := true
	for _, r := range data {
		cursor := algorithm(string(r))
		if cursor == "" {
			continue
		}

		isCursorDash = cursor == "-"
		isStartedDash = isCursorDash && isResultEmpty
		isDoubledDash = isCursorDash && !isResultEmpty && isPreviousDash

		if isStartedDash || isDoubledDash {
			continue
		}

		result = append(result, []rune(cursor)...)
		isPreviousDash = isCursorDash
		isResultEmpty = false
	}

	isEndedDash := !isResultEmpty && isCursorDash
	if isEndedDash {
		return string(result[:len(result)-1])
	}
	return string(result)
}

func algorithm(data string) string {
	var result string
	for ind := range data {
		char, ok := mapping[string([]rune(data)[ind])]
		if ok {
			result += char
		}
	}

	return result
}
