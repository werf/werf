package inspector

import (
	"fmt"
	"path/filepath"
)

func (i Inspector) InspectConfigDockerfileContextAddFile(relPath string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if i.giterminismConfig.IsConfigDockerfileContextAddFileAccepted(relPath) {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf(`contextAddFile %q not allowed by giterminism

The use of the directive contextAddFiles complicates the sharing and reproducibility of the configuration in CI jobs and among developers because the file data affects the final digest of built images and must be identical at all steps of the pipeline and during local development.`, filepath.ToSlash(relPath)))
}
