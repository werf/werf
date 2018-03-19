package main

import (
	"fmt"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("image", func(args map[string]interface{}) (map[string]interface{}, error) {
		fmt.Printf("image args: %v\n", args)

		res := make(map[string]interface{})
		res["id"] = "deadbeef"

		return res, nil
	})
}
