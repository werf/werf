package inspector

import (
	"context"

	"github.com/werf/werf/pkg/path_matcher"
)

func (i Inspector) InspectBuildContextFiles(ctx context.Context, matcher path_matcher.PathMatcher) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	return i.fileReader.ValidateStatusResult(ctx, matcher)
}
