package config

import (
	"fmt"
	"runtime"

	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/util/option"
)

type imagePlatformValidator struct{}

func newImagePlatformValidator() *imagePlatformValidator {
	return &imagePlatformValidator{}
}

// Validate that all dependencies/imports (image:platform) have required base images (image:platform)
// It respects
// a) base images syntaxes via:
// ---
// "image": test-from
// "from": <...> # (werf and external images)
// ---
// "image": test-from-image
// "fromImage": <...>
// ---
// "image": test-from-artifact
// "fromArtifact": <...>
// ---
// b) dependencies/imports syntaxes via:
// ---
// "image": test-dep
// dependencies:
//   - image: test-from
//
// ---
// image: test-import
// import:
//   - from: test-from
//     add: /usr/local/bin
//     to: /opt/test
//
// ---
func (v *imagePlatformValidator) Validate(rawStapelImages []*rawStapelImage, rawImagesFromDockerfile []*rawImageFromDockerfile) error {
	// Collect all combinations of image:platform from both sources
	allImagesPlatforms := make([]lo.Tuple2[string, string], 0, len(rawStapelImages)+len(rawImagesFromDockerfile))

	for _, img := range rawStapelImages {
		allImagesPlatforms = append(allImagesPlatforms, v.crossJoinImagesPlatforms(img.Images, img.Platform)...)
	}

	for _, img := range rawImagesFromDockerfile {
		allImagesPlatforms = append(allImagesPlatforms, v.crossJoinImagesPlatforms(img.Images, img.Platform)...)
	}

	for _, img := range rawStapelImages {
		for _, dep := range img.RawDependencies {
			_, rightDiff := lo.Difference(allImagesPlatforms, v.crossJoinImagesPlatforms([]string{dep.Image}, img.Platform))

			if len(rightDiff) > 0 {
				return v.newDependencyError(img.Images[0], rightDiff[0].A, rightDiff[0].B)
			}
		}

		for _, dep := range img.RawImport {
			_, rightDiff := lo.Difference(allImagesPlatforms, v.crossJoinImagesPlatforms([]string{option.ValueOrDefault(dep.From, dep.ImageName)}, img.Platform))

			if len(rightDiff) > 0 {
				return v.newImportError(img.Images[0], rightDiff[0].A, rightDiff[0].B)
			}
		}
	}

	for _, img := range rawImagesFromDockerfile {
		for _, dep := range img.RawDependencies {
			_, rightDiff := lo.Difference(allImagesPlatforms, v.crossJoinImagesPlatforms([]string{dep.Image}, img.Platform))

			if len(rightDiff) > 0 {
				return v.newDependencyError(img.Images[0], rightDiff[0].A, rightDiff[0].B)
			}
		}
	}

	return nil
}

func (v *imagePlatformValidator) crossJoinImagesPlatforms(names, platforms []string) []lo.Tuple2[string, string] {
	if len(platforms) == 0 {
		platforms = []string{fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)} // default platform
	}
	return lo.CrossJoin2(names, platforms)
}

func (v *imagePlatformValidator) newDependencyError(imageName, requiredImageName, requiredImagePlatform string) error {
	return fmt.Errorf("image=%q platform=%q requires dependency image=%q platform=%q which is not present in configuration", imageName, requiredImagePlatform, requiredImageName, requiredImagePlatform)
}

func (v *imagePlatformValidator) newImportError(imageName, requiredImageName, requiredImagePlatform string) error {
	return fmt.Errorf("image=%q platform=%q requires import image=%q platform=%q which is not present in configuration", imageName, requiredImagePlatform, requiredImageName, requiredImagePlatform)
}
