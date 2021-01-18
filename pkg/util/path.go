package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar"
)

func ExpandPath(path string) string {
	var result string

	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}

		dir := usr.HomeDir

		if path == "~" {
			result = dir
		} else {
			result = filepath.Join(dir, path[2:])
		}
	} else {
		var err error
		result, err = filepath.Abs(path)
		if err != nil {
			panic(err) // stupid interface of filepath.Abs
		}
	}

	return result
}

func SplitFilepath(path string) (result []string) {
	path = filepath.FromSlash(path)
	separator := os.PathSeparator

	idx := 0
	if separator == '\\' {
		// if the separator is '\\', then we can just split...
		result = strings.Split(path, string(separator))
		idx = len(result)
	} else {
		// otherwise, we need to be careful of situations where the separator was escaped
		cnt := strings.Count(path, string(separator))
		if cnt == 0 {
			return []string{path}
		}

		result = make([]string, cnt+1)
		pathlen := len(path)
		separatorLen := utf8.RuneLen(separator)
		emptyEnd := false
		for start := 0; start < pathlen; {
			end := indexRuneWithEscaping(path[start:], separator)
			if end == -1 {
				emptyEnd = false
				end = pathlen
			} else {
				emptyEnd = true
				end += start
			}
			result[idx] = path[start:end]
			start = end + separatorLen
			idx++
		}

		// If the last rune is a path separator, we need to append an empty string to
		// represent the last, empty path component. By default, the strings from
		// make([]string, ...) will be empty, so we just need to increment the count
		if emptyEnd {
			idx++
		}
	}

	return result[:idx]
}

// Find the first index of a rune in a string,
// ignoring any times the rune is escaped using "\".
func indexRuneWithEscaping(s string, r rune) int {
	end := strings.IndexRune(s, r)
	if end == -1 {
		return -1
	}
	if end > 0 && s[end-1] == '\\' {
		start := end + utf8.RuneLen(r)
		end = indexRuneWithEscaping(s[start:], r)
		if end != -1 {
			end += start
		}
	}
	return end
}

// WalkByPattern supports ** in pattern and does not follow symlinks
func WalkByPattern(dir, pattern string, walkFunc func(path string, f os.FileInfo, err error) error) error {
	patternRelDir := filepath.Join(dir, pattern)
	patternComponents := SplitFilepath(pattern)

	// figure out how many components we don't need to glob because they're
	// just names without patterns - we'll use os.Lstat below to check if that
	// path actually exists
	patLen := len(patternComponents)
	patIdx := 0
	for ; patIdx < patLen; patIdx++ {
		if strings.IndexAny(patternComponents[patIdx], "*?[{\\") >= 0 {
			break
		}
	}
	if patIdx > 0 {
		dir = filepath.Join(append([]string{dir}, patternComponents[0:patIdx]...)...)
	}

	if exist, err := FileExists(dir); err != nil {
		return err
	} else if !exist {
		return nil
	}

	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return walkFunc(path, f, err)
		}

		matched, err := doublestar.PathMatch(patternRelDir, path)
		if err != nil {
			return walkFunc(path, f, fmt.Errorf("path match failed: %s", err))
		}

		if matched {
			return walkFunc(path, f, nil)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("filepath walk failed: %s", err)
	}

	return nil
}
