package main

import (
	"fmt"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("git-artifact", func(args map[string]interface{}) (map[string]interface{}, error) {
		fmt.Printf("git-artifact args: %v\n", args)

		res := make(map[string]interface{})
		res["param"] = 300
		res["param2"] = "helo"

		return res, fmt.Errorf("sukablyat")
	})
}
