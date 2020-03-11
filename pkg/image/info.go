package image

import "time"

type Info struct {
	Repository string
	Tag        string
	ID         string
	ParentID   string
	ImageName  string `json:"imageName"`

	Signature         string            `json:"signature"`
	Labels            map[string]string `json:"labels"`
	CreatedAtUnixNano int64             `json:"createdAtUnixNano"`
}

func (info *Info) CreatedAt() time.Time {
	return time.Unix(info.CreatedAtUnixNano/1000_000_000, info.CreatedAtUnixNano%1000_000_000)
}
