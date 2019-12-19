package util

import (
	"os/user"
	"path/filepath"
	"strings"
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
