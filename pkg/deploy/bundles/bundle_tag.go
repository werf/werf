package bundles

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

	tsWithTagVersion := nonSemverTagVersion(tag, now)
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

// nonSemverTagVersion returns 0.0.0-<unix time>-<tag slug>. The chart is pushed under its version, so
// a slug that would make the version longer than a tag can be is cut and suffixed with a hash of the
// tag to tell such tags apart.
func nonSemverTagVersion(tag string, now time.Time) string {
	prefix := fmt.Sprintf("0.0.0-%d-", now.Unix())
	tagSlug := slug.Slug(tag)

	if len(prefix)+len(tagSlug) > slug.DockerTagMaxSize {
		hash := sha256.Sum256([]byte(tag))
		hashSuffix := "-" + hex.EncodeToString(hash[:])[:8]
		tagSlug = strings.TrimRight(tagSlug[:slug.DockerTagMaxSize-len(prefix)-len(hashSuffix)], "-") + hashSuffix
	}

	return prefix + tagSlug
}

// chartTags returns the tags to push a chart with chartVersion under: the version itself and the tag the user asked for, if it differs, both with "+", invalid in a tag, encoded as Helm does.
func chartTags(chartVersion, tag string) []string {
	versionTag := encodeTag(chartVersion)
	tag = encodeTag(tag)
	if tag == "" || tag == versionTag {
		return []string{versionTag}
	}

	return []string{versionTag, tag}
}

func encodeTag(tag string) string {
	return strings.ReplaceAll(tag, "+", "_")
}
