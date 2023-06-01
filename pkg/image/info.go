package image

import (
	"fmt"
	"strings"
	"time"

	"github.com/werf/werf/pkg/util"
)

const (
	DockerHubRepositoryPrefix      = "docker.io/library/"
	IndexDockerHubRepositoryPrefix = "index.docker.io/library/"
)

type Info struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	// repo@sha256:digest
	RepoDigest string `json:"repoDigest"`

	OnBuild           []string          `json:"onBuild"`
	Env               []string          `json:"env"`
	ID                string            `json:"ID"`
	ParentID          string            `json:"parentID"`
	Labels            map[string]string `json:"labels"`
	Size              int64             `json:"size"`
	CreatedAtUnixNano int64             `json:"createdAtUnixNano"`

	IsIndex bool
	Index   []*Info
}

func (info *Info) GetDigest() string {
	return strings.TrimPrefix(info.RepoDigest, fmt.Sprintf("%s@", info.Repository))
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
	res := &Info{
		Name:              info.Name,
		Repository:        info.Repository,
		Tag:               info.Tag,
		RepoDigest:        info.RepoDigest,
		OnBuild:           util.CopyArr(info.OnBuild),
		Env:               util.CopyArr(info.Env),
		ID:                info.ID,
		ParentID:          info.ParentID,
		Labels:            util.CopyMap(info.Labels),
		Size:              info.Size,
		CreatedAtUnixNano: info.CreatedAtUnixNano,

		IsIndex: info.IsIndex,
	}

	for _, i := range info.Index {
		res.Index = append(res.Index, i.GetCopy())
	}

	return res
}

func (info *Info) LogName() string {
	if info.Name == "<none>:<none>" {
		return info.ID
	} else {
		return info.Name
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

func NormalizeRepository(repository string) (res string) {
	res = repository
	res = strings.TrimPrefix(res, IndexDockerHubRepositoryPrefix)
	res = strings.TrimPrefix(res, DockerHubRepositoryPrefix)
	return
}

// ExtractRepoDigest return repo@digest from the list.
func ExtractRepoDigest(inspectRepoDigests []string, repository string) string {
	for _, inspectRepoDigest := range inspectRepoDigests {
		repoAndDigest := strings.SplitN(inspectRepoDigest, "@sha256:", 2)
		repo := NormalizeRepository(repoAndDigest[0])
		if len(repoAndDigest) == 2 && NormalizeRepository(repository) == repo {
			return fmt.Sprintf("%s@sha256:%s", repo, repoAndDigest[1])
		}
	}
	return ""
}
