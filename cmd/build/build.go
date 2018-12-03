package main

import "fmt"

type buildRubyCliOptions struct {
	Name     string   `json:"name"`
	BuildDir string   `json:"build_dir"`
	SSHKey   []string `json:"ssh_key"`
}

func runBuild(projectDir string, rubyCliOptions buildRubyCliOptions) error {
	fmt.Printf("runBuild called: %s %#v\n", projectDir, rubyCliOptions)
	return nil
}
