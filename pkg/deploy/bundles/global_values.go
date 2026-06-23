package bundles

import (
	"fmt"

	"github.com/werf/werf/v2/pkg/image"
)

func updateGlobalWerfValues(values map[string]interface{}, newRepo string, newImageRefs map[string]string) error {
	globalVals, ok := values["global"].(map[string]interface{})
	if !ok {
		return nil
	}

	werfVals, ok := globalVals["werf"].(map[string]interface{})
	if !ok {
		return nil
	}

	werfVals["repo"] = newRepo

	imagesVals, ok := werfVals["images"].(map[string]interface{})
	if !ok {
		return nil
	}

	for imageName, imageVals := range imagesVals {
		vals, ok := imageVals.(map[string]interface{})
		if !ok {
			continue
		}

		newRef, hasRef := newImageRefs[imageName]
		if !hasRef {
			continue
		}

		digest, _ := vals["digest"].(string)

		rebuilt, err := image.RebuildImageValuesMap(newRef, digest)
		if err != nil {
			return fmt.Errorf("update global werf image values for %q: %w", imageName, err)
		}

		imagesVals[imageName] = rebuilt
	}

	return nil
}
