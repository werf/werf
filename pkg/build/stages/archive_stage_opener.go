package stages

import (
	"context"
	"io"
)

type StageArchiveOpener struct {
	Archive  *ArchiveStorage
	ImageTag string
	ctx      context.Context
}

func NewStageArchiveOpener(archive *ArchiveStorage, imageTag string) *StageArchiveOpener {
	return &StageArchiveOpener{
		Archive:  archive,
		ImageTag: imageTag,
	}
}

func (opener *StageArchiveOpener) Open() (io.ReadCloser, error) {
	return opener.Archive.Reader.ReadArchiveStage(opener.ctx, opener.ImageTag)
}

func (opener *StageArchiveOpener) SetContext(ctx context.Context) {
	opener.ctx = ctx
}
