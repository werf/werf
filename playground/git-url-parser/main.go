package main

import (
	"fmt"

	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

func main() {
	for _, url := range []string{
		"git@github.com:flant/werf.git",
		"https://github.com/flant/werf.git",
		"https://aloha:privet@github.com/flant/werf.git",
		"/home/myuser/go-workspace/src/github.com/flant/werf/",
	} {
		ep, err := transport.NewEndpoint(url)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Endpoint:\n%#v\n---\n", ep)
	}
}
