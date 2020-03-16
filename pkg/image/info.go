package image

import "time"

type Info struct {
	Name       string
	Repository string
	Tag        string
	Digest     string
	ID         string
	ParentID   string
	Labels     map[string]string

	createdAtUnixNano int64
}

func (info *Info) SetCreatedAtUnix(seconds int64) {
	info.createdAtUnixNano = seconds * 1000_000_000
}

func (info *Info) SetCreatedAtUnixNano(seconds int64) {
	info.createdAtUnixNano = seconds
}

func (info *Info) CreatedAt() time.Time {
	return time.Unix(info.createdAtUnixNano/1000_000_000, info.createdAtUnixNano%1000_000_000)
}

func (info *Info) CreatedAtUnixNano() int64 {
	return info.createdAtUnixNano
}
