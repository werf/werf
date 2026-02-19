package image

import (
	"fmt"
	"strings"

	"github.com/distribution/reference"
)

const scratchImageName = "scratch"

func IsScratchRef(imageRef string) bool {
	if imageRef == "" {
		return false
	}

	if imageRef == scratchImageName {
		return true
	}

	ref, err := reference.ParseAnyReference(imageRef)
	if err != nil {
		return false
	}

	named, ok := ref.(reference.Named)
	if !ok {
		return false
	}

	path := reference.Path(named)

	return path == scratchImageName || strings.HasSuffix(path, "/"+scratchImageName)
}

func ImageName(name string) string {
	return fmt.Sprintf("%s-sbom", name)
}

func BaseImageName(repo, tag string) string {
	return ImageName(fmt.Sprintf("%s:%s", repo, tag))
}
