package slug

import (
	"os"

	"github.com/flant/dapp/pkg/slug/v1"
	"github.com/flant/dapp/pkg/slug/v2"
	"github.com/flant/dapp/pkg/slug/v3"
)

func Slug(data string) string {
	if os.Getenv("DAPP_SLUG_V3") != "" {
		return v3.Slug(data)
	} else if os.Getenv("DAPP_SLUG_V2") != "" {
		return v2.Slug(data)
	} else {
		return v1.Slug(data)
	}
}
