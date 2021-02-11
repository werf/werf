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

	return NewExternalDependencyFoundError(fmt.Sprintf("contextAddFile %q not allowed", filepath.ToSlash(relPath)))
}
