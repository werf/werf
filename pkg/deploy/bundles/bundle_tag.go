package bundles

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/slug"
)

var strictSemverTagRegexp = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func BundleTagToChartVersion(ctx context.Context, tag string, now time.Time) (*semver.Version, error) {
	if strictSemverTagRegexp.MatchString(tag) {
		sv, err := semver.NewVersion(tag)
		if err != nil {
			return nil, fmt.Errorf("parse semver tag %q: %w", tag, err)
		}

		return sv, nil
	}

	// TODO: come up with a better idea for generating reproducible, consistent and monotonously increasing semver

	tsWithTagVersion := fmt.Sprintf("0.0.0-%d-%s", now.Unix(), slug.Slug(tag))
	sv, err := semver.NewVersion(tsWithTagVersion)
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

// chartTags returns the tags to push a chart with chartVersion under: the version itself and the tag the user asked for, if it differs, both with "+", invalid in a tag, encoded as Helm does.
func chartTags(chartVersion, tag string) []string {
	versionTag := strings.ReplaceAll(chartVersion, "+", "_")
	tag = strings.ReplaceAll(tag, "+", "_")
	if tag == "" || tag == versionTag {
		return []string{versionTag}
	}

	return []string{versionTag, tag}
}
