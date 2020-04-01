package image

import (
	"path/filepath"

	"github.com/flant/werf/pkg/werf"
)

var CommonManifestCache *ManifestCache

func Init() error {
	CommonManifestCache = NewManifestCache(filepath.Join(werf.GetLocalCacheDir(), "manifest_cache", ManifestCacheVersion))
	return nil
}
