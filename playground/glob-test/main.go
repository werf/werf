package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
)

func main() {
	matches, err := doublestar.Glob("nofile")
	if err != nil {
		panic(err)
	}
	for _, p := range matches {
		fmt.Printf("-> %s\n", p)
	}

	err = filepath.Walk(".", func(walkPath string, info os.FileInfo, accessErr error) error {
		if accessErr != nil {
			return fmt.Errorf("error accessing `%s`: %s", walkPath, accessErr)
		}

		path, err := filepath.Rel(".", walkPath)
		if err != nil {
			return err
		}

		fmt.Printf("-> %s\n", path)

		return nil
	})

	if err != nil {
		panic(err)
	}
}
