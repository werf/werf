package image

import (
	"path/filepath"

	"github.com/werf/werf/pkg/werf"
)

var CommonManifestCache *ManifestCache

func Init() error {
	CommonManifestCache = NewManifestCache(filepath.Join(werf.GetLocalCacheDir(), "manifests", ManifestCacheVersion))
	return nil
}
