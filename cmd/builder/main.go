package main

import (
	"fmt"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("builder", func(args map[string]interface{}) (interface{}, error) {
		fmt.Printf("builder args: %v\n", args)

		return nil, fmt.Errorf("Some error!")
	})
}
