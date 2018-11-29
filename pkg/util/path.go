package util

import "path/filepath"

func ExpandPath(path string) string {
	res, err := filepath.Abs(path)
	if err != nil {
		panic(err) // stupid interface of filepath.Abs
	}
	return res
}
