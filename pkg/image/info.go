package image

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"github.com/werf/werf/pkg/util"
)

type Info struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	RepoDigest string `json:"repoDigest"`

	OnBuild           []string          `json:"onBuild"`
	ID                string            `json:"ID"`
	ParentID          string            `json:"parentID"`
	Labels            map[string]string `json:"labels"`
	Size              int64             `json:"size"`
	CreatedAtUnixNano int64             `json:"createdAtUnixNano"`
}

func (info *Info) SetCreatedAtUnix(seconds int64) {
	info.CreatedAtUnixNano = seconds * 1000_000_000
}

func (info *Info) SetCreatedAtUnixNano(seconds int64) {
	info.CreatedAtUnixNano = seconds
}

func (info *Info) GetCreatedAt() time.Time {
	return time.Unix(info.CreatedAtUnixNano/1000_000_000, info.CreatedAtUnixNano%1000_000_000)
}

func (info *Info) GetCopy() *Info {
	return &Info{
		Name:              info.Name,
		Repository:        info.Repository,
		Tag:               info.Tag,
		RepoDigest:        info.RepoDigest,
		OnBuild:           util.CopyArr(info.OnBuild),
		ID:                info.ID,
		ParentID:          info.ParentID,
		Labels:            util.CopyMap(info.Labels),
		Size:              info.Size,
		CreatedAtUnixNano: info.CreatedAtUnixNano,
	}
}

func NewInfoFromInspect(ref string, inspect *types.ImageInspect) *Info {
	repository, tag := ParseRepositoryAndTag(ref)

	var repoDigest string
	if len(inspect.RepoDigests) > 0 {
		// NOTE: suppose we have a single repo for each stage
		repoDigest = inspect.RepoDigests[0]
	}

	var parentID string
	if id, ok := inspect.Config.Labels["werf.io/base-image-id"]; ok {
		parentID = id
	} else {
		// TODO(1.3): Legacy compatibility mode
		parentID = inspect.Config.Image
	}

	return &Info{
		Name:              ref,
		Repository:        repository,
		Tag:               tag,
		Labels:            inspect.Config.Labels,
		OnBuild:           inspect.Config.OnBuild,
		CreatedAtUnixNano: MustParseTimestampString(inspect.Created).UnixNano(),
		RepoDigest:        repoDigest,
		ID:                inspect.ID,
		ParentID:          parentID,
		Size:              inspect.Size,
	}
}

func MustParseTimestampString(timestampString string) time.Time {
	t, err := time.Parse(time.RFC3339, timestampString)
	if err != nil {
		panic(fmt.Sprintf("got bad timestamp %q: %s", timestampString, err))
	}
	return t
}

func ParseRepositoryAndTag(ref string) (string, string) {
	parts := strings.SplitN(util.Reverse(ref), ":", 2)
	if len(parts) != 2 {
		return ref, ""
	}
	tag := util.Reverse(parts[0])
	repository := util.Reverse(parts[1])
	return repository, tag
}
