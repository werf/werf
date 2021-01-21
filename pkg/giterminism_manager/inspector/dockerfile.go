package inspector

import (
	"fmt"
	"path/filepath"
)

func (i Inspector) InspectConfigDockerfileContextAddFile(relPath string) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if isAccepted, err := i.giterminismConfig.IsConfigDockerfileContextAddFileAccepted(relPath); err != nil {
		return err
	} else if isAccepted {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf("contextAddFile %q not allowed", filepath.ToSlash(relPath)))
}
