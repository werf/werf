package signver

import "github.com/werf/werf/v2/pkg/signver/pass"

type KeyOpts struct {
	KeyRef   string
	PassFunc pass.PassFunc
}
