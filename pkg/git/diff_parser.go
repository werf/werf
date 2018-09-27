package git

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func makeDiffParser(out io.Writer, pathFilter PathFilter) *diffParser {
	return &diffParser{
		Out:        out,
		OutLines:   0,
		PathFilter: pathFilter,
		state:      unrecognized,
		lineBuf:    make([]byte, 0, 4096),
	}
}

type parserState string

const (
	unrecognized       parserState = "unrecognized"
	diffBegin          parserState = "diffBegin"
	diffBody           parserState = "diffBody"
	newFileDiff        parserState = "newFileDiff"
	deleteFileDiff     parserState = "deleteFileDiff"
	modifyFileDiff     parserState = "modifyFileDiff"
	modifyFileModeDiff parserState = "modifyFileModeDiff"
	ignoreDiff         parserState = "ignoreDiff"
)

type diffParser struct {
	PathFilter PathFilter

	Out                 io.Writer
	OutLines            uint
	UnrecognizedCapture bytes.Buffer

	state   parserState
	lineBuf []byte
}

func (p *diffParser) HandleStdout(data []byte) error {
	for _, b := range data {
		if b == '\n' {
			line := string(p.lineBuf)
			p.lineBuf = p.lineBuf[:0]

			err := p.handleDiffLine(line)
			if err != nil {
				return fmt.Errorf("error parsing diff line: %s", err)
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

func (p *diffParser) handleDiffLine(line string) error {
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
			return p.handleModifyFileModeDiff(line)
		}
		if strings.HasPrefix(line, "index ") {
			return p.handleModifyFileDiff(line)
		}
		return fmt.Errorf("unexpected diff line in state `%s`: %#v", p.state, line)

	case modifyFileDiff:
		if strings.HasPrefix(line, "--- ") {
			return p.handleModifyFilePathA(line)
		}
		if strings.HasPrefix(line, "+++ ") {
			return p.handleModifyFilePathB(line)
		}
		if strings.HasPrefix(line, "GIT binary patch") {
			return p.handleBinaryBeginHeader(line)
		}
		return p.writeOutLine(line)

	case modifyFileModeDiff:
		if strings.HasPrefix(line, "new mode ") {
			p.state = unrecognized
			return p.writeOutLine(line)
		}
		return fmt.Errorf("unexpected diff line in state `%s`: %#v", p.state, line)

	case newFileDiff:
		if strings.HasPrefix(line, "+++ ") {
			return p.handleNewFilePath(line)
		}
		if strings.HasPrefix(line, "GIT binary patch") {
			return p.handleBinaryBeginHeader(line)
		}
		return p.writeOutLine(line)

	case deleteFileDiff:
		if strings.HasPrefix(line, "--- ") {
			return p.handleDeleteFilePath(line)
		}
		if strings.HasPrefix(line, "GIT binary patch") {
			return p.handleBinaryBeginHeader(line)
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

func (p *diffParser) handleDiffBegin(line string) error {
	lineParts := strings.Split(line, " ")
	pathA := strings.TrimPrefix(lineParts[2], "a/")
	pathB := strings.TrimPrefix(lineParts[3], "b/")

	if !p.PathFilter.IsFilePathValid(pathA) || !p.PathFilter.IsFilePathValid(pathB) {
		p.state = ignoreDiff
		return nil
	}

	newPathA := p.PathFilter.TrimFileBasePath(pathA)
	newPathB := p.PathFilter.TrimFileBasePath(pathB)
	newLine := fmt.Sprintf("diff --git a/%s b/%s", newPathA, newPathB)

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

func (p *diffParser) handleModifyFileModeDiff(line string) error {
	p.state = modifyFileModeDiff
	return p.writeOutLine(line)
}

func (p *diffParser) handleModifyFilePathA(line string) error {
	path := strings.TrimPrefix(line, "--- a/")
	newPath := p.PathFilter.TrimFileBasePath(path)
	newLine := fmt.Sprintf("--- a/%s", newPath)
	return p.writeOutLine(newLine)
}

func (p *diffParser) handleModifyFilePathB(line string) error {
	path := strings.TrimPrefix(line, "+++ b/")
	newPath := p.PathFilter.TrimFileBasePath(path)
	newLine := fmt.Sprintf("+++ b/%s", newPath)

	p.state = diffBody

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleSubmoduleLine(line string) error {
	p.state = unrecognized
	return nil
}

func (p *diffParser) handleNewFilePath(line string) error {
	path := strings.TrimPrefix(line, "+++ b/")
	newPath := p.PathFilter.TrimFileBasePath(path)
	newLine := fmt.Sprintf("+++ b/%s", newPath)

	p.state = diffBody

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleDeleteFilePath(line string) error {
	path := strings.TrimPrefix(line, "--- a/")
	newPath := p.PathFilter.TrimFileBasePath(path)
	newLine := fmt.Sprintf("--- a/%s", newPath)

	p.state = diffBody

	return p.writeOutLine(newLine)
}

func (p *diffParser) handleBinaryBeginHeader(line string) error {
	p.state = diffBody
	return p.writeOutLine(line)
}
