package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"

	"github.com/flant/dapp/pkg/config"
)

var (
	WorkingDir      string
	DappProjectRoot string
)

func fprintResponse(w io.Writer, response map[string]string) {
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot dump response %v\n", response)
		os.Exit(1)
	}
	fmt.Fprintf(w, "%s\n", string(data))
}

func usage() {
	fmt.Fprintf(os.Stderr, "dappfile-yml\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	WorkingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine working dir: %s\n", err)
		os.Exit(1)
	}

	flag.Usage = usage
	flag.StringVar(&DappProjectRoot, "dapp-project-root", WorkingDir, "Directory where dappfile.yml resides")
	flag.Parse()

	var dappfilePath string
	for _, file := range []string{"dappfile.yml", "dappfile.yaml"} {
		checkPath := filepath.Join(DappProjectRoot, file)
		if _, err := os.Stat(checkPath); !os.IsNotExist(err) {
			dappfilePath = checkPath
			break
		}
	}
	if dappfilePath == "" {
		fprintResponse(os.Stderr, map[string]string{
			"error":   "dappfile_not_found",
			"message": fmt.Sprintf("dappfile.yml or dappfile.yaml is not found in %s", DappProjectRoot),
		})
		os.Exit(16)
	}

	config, err := config.LoadDappfile(dappfilePath)
	if err != nil {
		fprintResponse(os.Stderr, map[string]string{
			"error":   "bad_dappfile",
			"message": fmt.Sprintf("Bad dappfile %s: %s", dappfilePath, err),
		})
		os.Exit(16)
	}

	serializedConfig, err := yaml.Marshal(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot dump dappConfig yaml data: %s\n", err)
		os.Exit(1)
	}

	fprintResponse(os.Stdout, map[string]string{"dappConfig": string(serializedConfig)})
}
