package main

import (
	"github.com/flant/dapp/pkg/ruby2go"
	"github.com/flant/dapp/pkg/slug"
)

func main() {
	ruby2go.RunCli("slug", func(args map[string]interface{}) (interface{}, error) {
		value, err := ruby2go.StringOptionFromArgs("data", args)
		if err != nil {
			return err, nil
		}

		return slug.Slug(value), nil
	})
}
