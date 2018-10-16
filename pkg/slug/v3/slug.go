package v3

import (
	"strings"

	"github.com/flant/dapp/pkg/slug/v1"
	"github.com/flant/dapp/pkg/slug/v2"
	"github.com/flant/dapp/pkg/util"
)

func Slug(data string) string {
	if ShouldNotBeSlugged(data) {
		return data
	}

	sluggedData := v1.Slugify(data)
	murmurHash := util.MurmurHash(data)

	var slugParts []string
	if sluggedData != "" {
		croppedSluggedData := v2.CropSluggedData(sluggedData)
		if strings.HasPrefix(croppedSluggedData, "-") {
			slugParts = append(slugParts, croppedSluggedData[:len(croppedSluggedData)-1])
		} else {
			slugParts = append(slugParts, croppedSluggedData)
		}
	}
	slugParts = append(slugParts, murmurHash)

	consistentUniqSlug := strings.Join(slugParts, v1.SlugSeparator)

	return consistentUniqSlug
}

func ShouldNotBeSlugged(data string) bool {
	return v1.ShouldNotBeSlugged(data) && len(data) < v2.SlugLimitLength
}
