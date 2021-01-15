package inspector

import (
	"fmt"
)

func (i Inspector) InspectConfigStapelGitBranch() error {
	if i.manager.LooseGiterminism() || i.manager.Config().IsConfigStapelGitBranchAccepted() {
		return nil
	}

	return NewExternalDependencyFoundError("git branch directive not allowed")
}
