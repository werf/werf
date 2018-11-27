package main

import (
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/build"
)

func main() {
	c := build.Conveyor{}

	err := c.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR building: %s\n", err)
	}
}
