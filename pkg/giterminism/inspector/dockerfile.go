package inspector

import (
	"fmt"
	"path/filepath"
)

func (i Inspector) InspectConfigDockerfileContextAddFile(relPath string) error {
	if i.manager.LooseGiterminism() {
		return nil
	}

	if isAccepted, err := i.manager.Config().IsConfigDockerfileContextAddFileAccepted(relPath); err != nil {
		return err
	} else if isAccepted {
		return nil
	}

	return NewExternalDependencyFoundError(fmt.Sprintf("contextAddFile '%s' not allowed", filepath.ToSlash(relPath)))
}
