package image

import (
	"path/filepath"

	"github.com/werf/werf/v2/pkg/werf"
)

var CommonManifestCache *ManifestCache

func Init() error {
	CommonManifestCache = NewManifestCache(filepath.Join(werf.GetLocalCacheDir(), "manifests", ManifestCacheVersion))
	return nil
}
