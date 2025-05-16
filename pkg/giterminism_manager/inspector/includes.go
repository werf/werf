package inspector

import "github.com/werf/werf/v2/pkg/giterminism_manager/errors"

var errIncludesUpdate = errors.NewError(`using latest versions of includes is not allowed by giterminism

The use of --includes-update option might make builds and deployments unreproducible.`)

func (i Inspector) InspectIncludesAllowUpdate() (bool, error) {
	if accepted := i.giterminismConfig.IsUpdateIncludesAccepted(); !accepted {
		return false, errIncludesUpdate
	}

	return true, nil
}
