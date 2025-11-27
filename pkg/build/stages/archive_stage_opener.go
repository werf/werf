package stages

import "io"

type StageArchiveOpener struct {
	Archive  *ArchiveStorage
	ImageTag string
}

func NewStageArchiveOpener(archive *ArchiveStorage, imageTag string) *StageArchiveOpener {
	return &StageArchiveOpener{
		Archive:  archive,
		ImageTag: imageTag,
	}
}

func (opener *StageArchiveOpener) Open() (io.ReadCloser, error) {
	return opener.Archive.Reader.ReadArchiveStage(opener.ImageTag)
}
