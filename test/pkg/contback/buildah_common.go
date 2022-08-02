package contback

import (
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
)

type BuildahInspect struct {
	Docker struct {
		Config manifest.Schema2Config `json:"config"`
	} `json:"Docker"`
}
