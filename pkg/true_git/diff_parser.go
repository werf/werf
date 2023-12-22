package true_git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

var ErrCommitsNotPresent = errors.New("commits not present")

func IsCommitsNotPresentError(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), ErrCommitsNotPresent.Error())
}

func makeDiffParser(out io.Writer, pathScope string, pathMatcher path_matcher.PathMatcher, fileRenames map[string]string) *diffParser {
	return &diffParser{
		PathScope:   pathScope,
		PathMatcher: pathMatcher,
		FileRenames: fileRenames,
		Out:         out,
		OutLines:    0,
		Paths:       make([]string, 0),
		BinaryPaths: make([]string, 0),
		state:       unrecognized,
		lineBuf:     make([]byte, 0, 4096),
	}
}

type parserState string

const (
	unrecognized   parserState = "unrecognized"
	diffBegin      parserState = "diffBegin"
	diffBody       parserState = "diffBody"
	newFileDiff    parserState = "newFileDiff"
	deleteFileDiff parserState = "deleteFileDiff"
	modifyFileDiff parserState = "modifyFileDiff"
	ignoreDiff     parserState = "ignoreDiff"
)

type diffParser struct {
	PathScope   string // Determines the directory that will get into the result (similar to <pathspec> in the git commands).
	PathMatcher path_matcher.PathMatcher
	FileRenames map[string]string // Files to rename during patching. Git repo relative paths of original files as keys, new filenames (without base path) as values.

	Out                 io.Writer
	OutLines            uint
	UnrecognizedCapture bytes.Buffer

	Paths         []string
	BinaryPaths   []string
	PathsToRemove []string
	LastSeenPaths []string

	state   parserState
	lineBuf []byte
}

func appendUnique(list []string, value string) []string {
	for _, oldValue := range list {
		if value == oldValue {
			return list
		}
	}
	return append(list, value)
}

func (p *diffParser) HandleStdout(data []byte) error {
	for _, b := range data {
		if b == '\n' {
			line := string(p.lineBuf)
			p.lineBuf = p.lineBuf[:0]

			err := p.handleDiffLine(line)
			if err != nil {
				return fmt.Errorf("error parsing diff line: %w", err)
			}

			continue
		}
		p.lineBuf = append(p.lineBuf, b)
	}

	return nil
}

func (p *diffParser) HandleStderr(data []byte) error {
	_, err := p.UnrecognizedCapture.Write(data)
	return err
}

func (p *diffParser) writeOutLine(line string) error {
	p.OutLines++

	_, err := p.Out.Write([]byte(line + "\n"))

	return err
}

func (p *diffParser) writeUnrecognizedLine(line string) error {
	_, err := p.UnrecognizedCapture.Write([]byte(line + "\n"))
	return err
}

func debugPatchParser() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_PATCH_PARSER") == "1"
}

func (p *diffParser) handleDiffLine(line string) error {
	if debugPatchParser() {
		oldState := p.state
		fmt.Printf("TRUE_GIT parse diff line: state=%#v line=%#v\n", oldState, line)
		defer func() {
			fmt.Printf("TRUE_GIT parse diff line: state change: %#v => %#v\n", oldState, p.state)
		}()
	}

	switch p.state {
	case unrecognized:
		if strings.HasPrefix(line, "diff --git ") {
			return p.handleDiffBegin(line)
		}
		if strings.HasPrefix(line, "Submodule ") {
			return p.handleSubmoduleLine(line)
		}
		return p.writeUnrecognizedLine(line)

	case ignoreDiff:
		if strings.HasPrefix(line, "diff --git ") {
			return p.handleDiffBegin(line)
		}
		if strings.HasPrefix(line, "Submodule ") {
			return p.handleSubmoduleLine(line)
		}
		return nil

	case diffBegin:
		if strings.HasPrefix(line, "deleted file mode ") {
			return p.handleDeleteFileDiff(line)
		}
		if strings.HasPrefix(line, "new file mode ") {
			return p.handleNewFileDiff(line)
		}
		if strings.HasPrefix(line, "old mode ") {
			return p.handleModifyFileDiff(line)
		}
		if strings.HasPrefix(line, "index ") {
			p.state = modifyFileDiff
			return p.handleIndexDiffLine(line)
		}
		return fmt.Errorf("unexpected diff line in state `%s`: %#v", p.state, line)

	case modifyFileDiff:
		if strings.HasPrefix(line, "new mode ") {
			return p.writeOutLine(line)
		}
		if strings.HasPrefix(line, "--- ") {
			return p.handleModifyFilePathA(line)
		}
		if strings.HasPrefix(line, "+++ ") {
			return p.handleModifyFilePathB(line)
		}
		if strings.HasPrefix(line, "GIT binary patch") {
			return p.handleBinaryBeginHeader(line)
		}
		if strings.HasPrefix(line, "Binary files") {
			return p.handleShortBinaryHeader(line)
		}
		if strings.HasPrefix(line, "diff --git ") {
			return p.handleDiffBegin(line)
		}
		if strings.HasPrefix(line, "index ") {
			return p.handleIndexDiffLine(line)
		}
		if strings.HasPrefix(line, "Submodule ") {
			return p.handleSubmoduleLine(line)
		}
		return p.writeOutLine(line)

	case newFileDiff:
		if strings.HasPrefix(line, "+++ ") {
			return p.handleNewFilePath(line)
		}
		if strings.HasPrefix(line, "GIT binary patch") {
			return p.handleBinaryBeginHeader(line)
		}
		if strings.HasPrefix(line, "Binary files") {
			return p.handleShortBinaryHeader(line)
		}
		if strings.HasPrefix(line, "diff --git ") {
			return p.handleDiffBegin(line)
		}
		if strings.HasPrefix(line, "index ") {
			return p.handleIndexDiffLine(line)
		}
		if strings.HasPrefix(line, "Submodule ") {
			return p.handleSubmoduleLine(line)
		}
		return p.writeOutLine(line)

	case deleteFileDiff:
		if strings.HasPrefix(line, "--- ") {
			return p.handleDeleteFilePath(line)
		}
		if strings.HasPrefix(line, "GIT binary patch") {
			return p.handleBinaryBeginHeader(line)
		}
		if strings.HasPrefix(line, "Binary files") {
			return p.handleShortBinaryHeader(line)
		}
		if strings.HasPrefix(line, "diff --git ") {
			return p.handleDiffBegin(line)
		}
		if strings.HasPrefix(line, "index ") {
			return p.handleIndexDiffLine(line)
		}
		if strings.HasPrefix(line, "Submodule ") {
			return p.handleSubmoduleLine(line)
		}
		return p.writeOutLine(line)

	case diffBody:
		if strings.HasPrefix(line, "diff --git ") {
			return p.handleDiffBegin(line)
		}
		if strings.HasPrefix(line, "Submodule ") {
			return p.handleSubmoduleLine(line)
		}
		return p.writeOutLine(line)
	}

	return nil
}

func (p *diffParser) handleDiffBegin(line string) (err error) {
	var lineParts []string
	var aAndBParts []string
	var a, b string

	lineParts = strings.SplitN(line, " \"a/", 2)
	if len(lineParts) == 2 {
		aAndBParts = strings.SplitN(lineParts[1], " \"b/", 2)
		a, b = fmt.Sprintf("\"a/%s", aAndBParts[0]), fmt.Sprintf("\"b/%s", aAndBParts[1])
	} else {
		lineParts = strings.SplitN(line, " a/", 2)
		if len(lineParts) == 2 {
			aAndBParts = strings.SplitN(lineParts[1], " b/", 2)
			a, b = fmt.Sprintf("a/%s", aAndBParts[0]), fmt.Sprintf("b/%s", aAndBParts[1])
		} else {
			return fmt.Errorf("unexpected diff begin line: %#v", line)
		}
	}

	trimmedPaths := make(map[string]string)

	p.LastSeenPaths = nil

	for _, data := range []struct{ PathWithPrefix, Prefix string }{{a, "a/"}, {b, "b/"}} {
		isPathQuoted := strings.HasPrefix(data.PathWithPrefix, "\"") && strings.HasSuffix(data.PathWithPrefix, "\"")

		var pathWithPrefix string
		if isPathQuoted {
			pathWithPrefix, err = strconv.Unquote(data.PathWithPrefix)
			if err != nil {
				return fmt.Errorf("unable to unquote diff path %#v: %w", data.PathWithPrefix, err)
			}
		} else {
			pathWithPrefix = data.PathWithPrefix
		}

		path := strings.TrimPrefix(pathWithPrefix, data.Prefix)
		if !p.PathMatcher.IsPathMatched(path) {
			p.state = ignoreDiff
			return nil
		}

		path = p.applyFileRenames(path)
		newPath := p.trimFileBaseFilepath(path)
		p.Paths = appendUnique(p.Paths, newPath)
		p.LastSeenPaths = appendUnique(p.LastSeenPaths, newPath)

		if isPathQuoted {
			trimmedPaths[data.PathWithPrefix] = strconv.Quote(data.Prefix + newPath)
		} else {
			trimmedPaths[data.PathWithPrefix] = data.Prefix + newPath
		}
	}

	newLine := fmt.Sprintf("diff --git %s %s", trimmedPaths[a], trimmedPaths[b])

	p.state = diffBegin

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleDeleteFileDiff(line string) error {
	p.state = deleteFileDiff
	return p.writeOutLine(line)
}

func (p *diffParser) handleNewFileDiff(line string) error {
	p.state = newFileDiff
	return p.writeOutLine(line)
}

func (p *diffParser) handleModifyFileDiff(line string) error {
	p.state = modifyFileDiff
	return p.writeOutLine(line)
}

// TODO: remove index line from resulting patch completely in v1.2
func (p *diffParser) handleIndexDiffLine(line string) error {
	var prefix, hashes, suffix string

	parts := strings.SplitN(line, " ", 3)
	switch {
	case len(parts) == 3:
		prefix, hashes, suffix = parts[0], parts[1], parts[2]
	case len(parts) == 2:
		prefix, hashes = parts[0], parts[1]
	default:
		return p.writeOutLine(line)
	}

	hashesParts := strings.SplitN(hashes, "..", 2)
	if len(hashesParts) != 2 {
		// unexpected format
		return p.writeOutLine(line)
	}

	stripHashFunc := func(h string) string {
		// TODO: remove index line from resulting patch completely in v1.2
		if len(h) < 8 {
			return h
		}
		return h[:8]
	}

	var leftHashes []string
	for _, h := range strings.Split(hashesParts[0], ",") {
		leftHashes = append(leftHashes, stripHashFunc(h))
	}

	var rightHashes []string
	for _, h := range strings.Split(hashesParts[1], ",") {
		rightHashes = append(rightHashes, stripHashFunc(h))
	}

	var newLine string

	if suffix == "" {
		newLine = fmt.Sprintf("%s %s..%s", prefix, strings.Join(leftHashes, ","), strings.Join(rightHashes, ","))
	} else {
		newLine = fmt.Sprintf("%s %s..%s %s", prefix, strings.Join(leftHashes, ","), strings.Join(rightHashes, ","), suffix)
	}

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleModifyFilePathA(line string) error {
	path := p.applyFileRenames(strings.TrimPrefix(line, "--- a/"))
	newPath := p.trimFileBaseFilepath(path)
	newLine := fmt.Sprintf("--- a/%s", newPath)

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleModifyFilePathB(line string) error {
	path := p.applyFileRenames(strings.TrimPrefix(line, "+++ b/"))
	newPath := p.trimFileBaseFilepath(path)
	newLine := fmt.Sprintf("+++ b/%s", newPath)

	p.state = diffBody

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleSubmoduleLine(line string) error {
	p.state = unrecognized
	if strings.HasSuffix(line, " (commits not present)") {
		return fmt.Errorf("cannot handle git diff line %q, check specified commits are correct: %w", line, ErrCommitsNotPresent)
	}
	return nil
}

func (p *diffParser) handleNewFilePath(line string) error {
	path := p.applyFileRenames(strings.TrimPrefix(line, "+++ b/"))
	newPath := p.trimFileBaseFilepath(path)
	newLine := fmt.Sprintf("+++ b/%s", newPath)

	p.state = diffBody

	return p.writeOutLine(newLine)
}

func (p *diffParser) trimFileBaseFilepath(path string) string {
	newPath := filepath.ToSlash(util.GetRelativeToBaseFilepath(filepath.FromSlash(p.PathScope), filepath.FromSlash(path)))
	// NOTE: for some files git diff may emit tabs on the end of path-line, for example: "--- a/path with spaces/to/file.txt\t"
	newPath = strings.TrimRight(newPath, "\t")
	return newPath
}

func (p *diffParser) handleDeleteFilePath(line string) error {
	path := p.applyFileRenames(strings.TrimPrefix(line, "--- a/"))
	newPath := p.trimFileBaseFilepath(path)
	newLine := fmt.Sprintf("--- a/%s", newPath)

	p.state = diffBody
	p.PathsToRemove = appendUnique(p.PathsToRemove, newPath)

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleBinaryBeginHeader(line string) error {
	for _, path := range p.LastSeenPaths {
		p.BinaryPaths = appendUnique(p.BinaryPaths, path)
	}

	p.state = diffBody

	return p.writeOutLine(line)
}

func (p *diffParser) handleShortBinaryHeader(line string) error {
	for _, path := range p.LastSeenPaths {
		p.BinaryPaths = appendUnique(p.BinaryPaths, path)
	}

	p.state = unrecognized

	return p.writeOutLine(line)
}

func (p *diffParser) applyFileRenames(path string) string {
	if renamedFileName, willRename := p.FileRenames[path]; willRename {
		return filepath.ToSlash(filepath.Join(p.PathScope, renamedFileName))
	}
	return path
}
