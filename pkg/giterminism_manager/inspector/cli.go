package inspector

import "github.com/werf/werf/pkg/giterminism_manager/errors"

func (i Inspector) InspectCustomTags() error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if i.giterminismConfig.IsCustomTagsAccepted() {
		return nil
	}

	return errors.NewError("the custom tags are not accepted")
}
