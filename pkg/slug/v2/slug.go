package v2

import (
	"strings"

	"github.com/flant/dapp/pkg/slug/v1"
	"github.com/flant/dapp/pkg/util"
)

const SlugLimitLength = 53

func Slug(data string) string {
	if v1.ShouldNotBeSlugged(data) {
		return data
	}

	sluggedData := v1.Slugify(data)
	murmurHash := util.MurmurHash(data)

	var slugParts []string
	if sluggedData != "" {
		croppedSluggedData := CropSluggedData(sluggedData)
		slugParts = append(slugParts, croppedSluggedData)
	}
	slugParts = append(slugParts, murmurHash)

	consistentUniqSlug := strings.Join(slugParts, v1.SlugSeparator)

	return consistentUniqSlug
}

func CropSluggedData(data string) string {
	var index int
	maxLength := SlugLimitLength - len(util.MurmurHash()) - len(v1.SlugSeparator)
	if len(data) > maxLength {
		index = maxLength
	} else {
		index = len(data)
	}

	return data[:index]
}
