package git_repo

import (
	"bytes"
	"fmt"
	"io"

	"github.com/flant/go-git/plumbing"
	"github.com/flant/go-git/plumbing/filemode"
	fdiff "github.com/flant/go-git/plumbing/format/diff"
	"github.com/flant/go-git/plumbing/object"
)

type RelativeFilteredPatch struct {
	PathFilter  PathFilter
	OriginPatch *object.Patch
}

func (p *RelativeFilteredPatch) FilePatches() []fdiff.FilePatch {
	res := make([]fdiff.FilePatch, 0)

	for _, fp := range p.OriginPatch.FilePatches() {
		from, to := fp.Files()

		if from != nil {
			if !p.PathFilter.IsFilePathValid(from.Path()) {
				continue
			}
		}

		if to != nil {
			if !p.PathFilter.IsFilePathValid(to.Path()) {
				continue
			}
		}

		res = append(res, &RelativeFilteredFilePatch{
			OriginFilePatch: fp,
			Patch:           p,
		})
	}

	return res
}

func (p *RelativeFilteredPatch) Message() string {
	return p.OriginPatch.Message()
}

func (p *RelativeFilteredPatch) Encode(w io.Writer) error {
	ue := fdiff.NewUnifiedEncoder(w, fdiff.DefaultContextLines)

	return ue.Encode(p)
}

func (p *RelativeFilteredPatch) String() string {
	buf := bytes.NewBuffer(nil)
	err := p.Encode(buf)
	if err != nil {
		panic(fmt.Sprintf("malformed patch: %s", err.Error()))
	}

	return buf.String()
}

type RelativeFilteredFilePatch struct {
	OriginFilePatch fdiff.FilePatch
	Patch           *RelativeFilteredPatch
}

func (fp *RelativeFilteredFilePatch) IsBinary() bool {
	return fp.OriginFilePatch.IsBinary()
}

func (fp *RelativeFilteredFilePatch) Files() (fdiff.File, fdiff.File) {
	var from, to fdiff.File

	fromF, toF := fp.OriginFilePatch.Files()

	if fromF != nil {
		from = &RelativeFilteredFile{OriginFile: fromF, FilePatch: fp}
	}

	if toF != nil {
		to = &RelativeFilteredFile{OriginFile: toF, FilePatch: fp}
	}

	return from, to
}

func (fp *RelativeFilteredFilePatch) Chunks() []fdiff.Chunk {
	return fp.OriginFilePatch.Chunks()
}

type RelativeFilteredFile struct {
	OriginFile fdiff.File
	FilePatch  *RelativeFilteredFilePatch
}

func (f *RelativeFilteredFile) Hash() plumbing.Hash {
	return f.OriginFile.Hash()
}

func (f *RelativeFilteredFile) Mode() filemode.FileMode {
	return f.OriginFile.Mode()
}

func (f *RelativeFilteredFile) GetAbsolutePath() string {
	return NormalizeAbsolutePath(f.OriginFile.Path())
}

func (f *RelativeFilteredFile) Path() string {
	if !f.FilePatch.Patch.PathFilter.IsFilePathValid(f.OriginFile.Path()) {
		panic(fmt.Errorf("failed assertion: patch path `%s` is not suitable for patch path filter %+v", f.OriginFile.Path(), f.FilePatch.Patch.PathFilter))
	}

	// f.OriginFile.Path() should be always a path to the file, not a directory
	return f.FilePatch.Patch.PathFilter.TrimFileBasePath(f.OriginFile.Path())
}
