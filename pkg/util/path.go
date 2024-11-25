package util

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"
)

// ExpandPath expands the path, replacing the tilde with the home directory and resolving the absolute path.
func ExpandPath(p string) (string, error) {
	p, err := ReplaceTildeWithHome(p)
	if err != nil {
		return p, err
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return p, err
	}

	return p, nil
}

func SplitFilepath(path string) (result []string) {
	separator := os.PathSeparator

	path = filepath.FromSlash(path)
	path = filepath.Clean(path)

	if runtime.GOOS == "windows" {
		sepStr := string(separator)
		uncRootPath := fmt.Sprintf("%s%s", sepStr, sepStr)

		path = strings.TrimPrefix(path, uncRootPath)
		path = strings.TrimPrefix(path, sepStr)
		path = strings.TrimSuffix(path, sepStr)
	} else if filepath.IsAbs(path) {
		p, err := filepath.Rel(string(separator), path)
		if err != nil {
			panic(fmt.Sprintf("unable to get relative path for %q", path))
		}
		path = p
	}

	if path == "" || path == "." || path == string(separator) {
		return nil
	}

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

// GlobPrefixWithoutPatterns figures out how many components we don't need to glob because they're
// just names without patterns
func GlobPrefixWithoutPatterns(glob string) (string, string) {
	var prefix string

	globComponents := SplitFilepath(glob)
	length := len(globComponents)
	ind := 0
	for ; ind < length; ind++ {
		if strings.ContainsAny(globComponents[ind], "*?[{\\") {
			break
		}
	}
	if ind > 0 {
		prefix = filepath.Join(globComponents[0:ind]...)
		glob = filepath.Join(globComponents[ind:]...)
	}

	return prefix, glob
}

func FilepathsWithParents(path string) []string {
	var res []string
	base := ""
	for _, part := range SplitFilepath(path) {
		base = filepath.Join(base, part)
		res = append(res, base)
	}

	return res
}

// SafeTrimGlobsAndSlashesFromFilepath trims any trailing globs and/or slashes from the path,
// while ignoring globs that are part of a directory or file name.
func SafeTrimGlobsAndSlashesFromFilepath(p string) string {
	return filepath.FromSlash(SafeTrimGlobsAndSlashesFromPath(p))
}

func SafeTrimGlobsAndSlashesFromPath(p string) string {
	parts := SplitFilepath(p)
	for i := len(parts) - 1; i >= 0; i-- {
		if partWOGlobs := strings.TrimRight(parts[i], "*"); partWOGlobs != "" {
			parts = parts[:i+1]
			break
		} else {
			parts = parts[:i]
		}
	}

	return path.Join(parts...)
}

func ReplaceTildeWithHome(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		firstSlash := strings.Index(path, "/")
		if firstSlash == 1 || firstSlash == -1 {
			home, err := os.UserHomeDir()
			if err != nil {
				return path, err
			}
			return strings.Replace(path, "~", home, 1), nil
		} else if firstSlash > 1 {
			username := path[1:firstSlash]
			userAccount, err := user.Lookup(username)
			if err != nil {
				return path, err
			}
			return strings.Replace(path, path[:firstSlash], userAccount.HomeDir, 1), nil
		}
	}

	return path, nil
}
