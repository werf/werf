package bundles

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/slug"
)

func BundleTagToChartVersion(ctx context.Context, tag string, now time.Time) (*semver.Version, error) {
	sv, err := semver.NewVersion(tag)
	if err == nil {
		return sv, nil
	}

	// TODO: come up with a better idea for generating reproducible, consistent and monotonously increasing semver

	tsWithTagVersion := fmt.Sprintf("0.0.0-%d-%s", now.Unix(), slug.Slug(tag))
	sv, err = semver.NewVersion(tsWithTagVersion)
	if err == nil {
		return sv, nil
	}

	fallbackVersion := fmt.Sprintf("0.0.0-%d", now.Unix())
	logboek.Context(ctx).Warn().LogF("Unable to use %q as chart version, will fallback on chart version %q\n", tsWithTagVersion, fallbackVersion)

	sv, err = semver.NewVersion(fallbackVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to use fallback chart version %q: %w", fallbackVersion, err)
	}

	return sv, nil
}
