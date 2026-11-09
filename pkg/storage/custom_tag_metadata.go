package storage

import (
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/slug"
)

// CustomTag is a custom tag together with the repository it was placed in.
// An empty Repo means the primary stages repository.
type CustomTag struct {
	Repo string
	Tag  string
}

func (t CustomTag) String() string {
	if t.Repo == "" {
		return t.Tag
	}
	return t.Repo + ":" + t.Tag
}

// metadataID is the record suffix in the meta repository. Tags of the primary
// repository keep the pre-Repo format so that records written by older werf
// versions stay addressable. Tags of other repositories are prefixed with a
// fixed-width hash of the canonical repository address: a plain join of
// repo and tag would be ambiguous ("app-install"+"main" vs "app"+"install-main").
func (t CustomTag) metadataID() string {
	if t.Repo == "" {
		return slug.LimitedSlug(t.Tag, 48)
	}
	repoHash := util.LegacyMurmurHash(canonicalRepoAddressOrRaw(t.Repo))
	return repoHash + "-" + slug.LimitedSlug(t.Tag, 48-len(repoHash)-1)
}

// SameRepoAddress compares two repository addresses canonically, so equivalent
// spellings (docker.io/x vs index.docker.io/library/x) are equal; an empty
// address never equals a non-empty one.
func SameRepoAddress(a, b string) bool {
	if a == "" || b == "" {
		return a == b
	}
	return canonicalRepoAddressOrRaw(a) == canonicalRepoAddressOrRaw(b)
}

func canonicalRepoAddressOrRaw(address string) string {
	canonical, err := CanonicalRepoAddress(address)
	if err != nil {
		return address
	}
	return canonical
}

type CustomTagMetadata struct {
	StageID string
	Tag     string
	Repo    string
}

func newCustomTagMetadata(stageID string, tag CustomTag) *CustomTagMetadata {
	return &CustomTagMetadata{
		StageID: stageID,
		Tag:     tag.Tag,
		Repo:    tag.Repo,
	}
}

func (m *CustomTagMetadata) CustomTag() CustomTag {
	return CustomTag{Repo: m.Repo, Tag: m.Tag}
}

func (m *CustomTagMetadata) ToLabels() map[string]string {
	labels := map[string]string{
		image.WerfCustomTagMetadataStageIDLabel: m.StageID,
		image.WerfCustomTagMetadataTag:          m.Tag,
	}
	if m.Repo != "" {
		labels[image.WerfCustomTagMetadataRepoLabel] = m.Repo
	}
	return labels
}

func newCustomTagMetadataFromLabels(labels map[string]string) *CustomTagMetadata {
	return &CustomTagMetadata{
		StageID: labels[image.WerfCustomTagMetadataStageIDLabel],
		Tag:     labels[image.WerfCustomTagMetadataTag],
		Repo:    labels[image.WerfCustomTagMetadataRepoLabel],
	}
}
