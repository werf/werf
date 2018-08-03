package git_repo

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	fdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type RelativeFilteredPatch struct {
	BasePath     string
	IncludePaths []string
	ExcludePaths []string
	OriginPatch  *object.Patch
}

func NormalizeAbsolutePath(path string) string {
	return filepath.Clean(filepath.Join("/", path))
}

func IsPathMatchesPattern(path, pattern string) bool {
	path = NormalizeAbsolutePath(path)
	pattern = NormalizeAbsolutePath(pattern)

	if strings.HasPrefix(path, pattern) {
		return true
	}

	matched, err := doublestar.PathMatch(pattern, path)
	if err != nil {
		panic(err)
	}
	if matched {
		return true
	}

	matched, err = doublestar.PathMatch(filepath.Join(pattern, "**", "*"), path)
	if err != nil {
		panic(err)
	}
	if matched {
		return true
	}

	return false
}

func IsPathMatchesOneOfPatterns(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if IsPathMatchesPattern(path, pattern) {
			return true
		}
	}

	return false
}

func (p *RelativeFilteredPatch) GetBasePath() string {
	return NormalizeAbsolutePath(p.BasePath)
}

func (p *RelativeFilteredPatch) IsInPath(path string) bool {
	return strings.HasPrefix(NormalizeAbsolutePath(path), p.GetBasePath())
}

func (p *RelativeFilteredPatch) GetRelativePath(path string) string {
	return strings.TrimPrefix(
		strings.TrimPrefix(NormalizeAbsolutePath(path), p.GetBasePath()),
		"/",
	)
}

func (p *RelativeFilteredPatch) FilePatches() []fdiff.FilePatch {
	res := make([]fdiff.FilePatch, 0)

	for _, fp := range p.OriginPatch.FilePatches() {
		from, to := fp.Files()

		if from != nil {
			if !p.IsInPath(from.Path()) {
				continue
			}

			if len(p.IncludePaths) > 0 {
				if !IsPathMatchesOneOfPatterns(from.Path(), p.IncludePaths) {
					continue
				}
			}

			if len(p.ExcludePaths) > 0 {
				if IsPathMatchesOneOfPatterns(from.Path(), p.ExcludePaths) {
					continue
				}
			}
		}

		if to != nil {
			if !p.IsInPath(to.Path()) {
				continue
			}

			if len(p.IncludePaths) > 0 {
				if !IsPathMatchesOneOfPatterns(to.Path(), p.IncludePaths) {
					continue
				}
			}

			if len(p.ExcludePaths) > 0 {
				if IsPathMatchesOneOfPatterns(to.Path(), p.ExcludePaths) {
					continue
				}
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
		return fmt.Sprintf("malformed patch: %s", err.Error())
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

func (f *RelativeFilteredFile) Path() string {
	return f.FilePatch.Patch.GetRelativePath(f.OriginFile.Path())
}
