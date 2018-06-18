package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/flant/yaml.v2"

	"github.com/flant/dapp/pkg/config"
)

var (
	WorkingDir   string
	DappfilePath string
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
	flag.StringVar(&DappfilePath, "dappfile", "", "Full path to dappfile.yml (dappfile.yml or dappfile.yaml from working directory will be used by default)")
	flag.Parse()

	if DappfilePath == "" {
		var defaultDappfilePath string
		for _, file := range []string{"dappfile.yml", "dappfile.yaml"} {
			checkPath := filepath.Join(WorkingDir, file)
			if _, err := os.Stat(checkPath); !os.IsNotExist(err) {
				defaultDappfilePath = checkPath
				break
			}
		}

		if defaultDappfilePath == "" {
			fprintResponse(os.Stderr, map[string]string{
				"error":   "dappfile_not_found",
				"message": fmt.Sprintf("dappfile.yml or dappfile.yaml is not found in %s", WorkingDir),
			})
			os.Exit(16)
		} else {
			DappfilePath = defaultDappfilePath
		}
	} else if _, err := os.Stat(DappfilePath); os.IsNotExist(err) {
		fprintResponse(os.Stderr, map[string]string{
			"error":   "dappfile_not_found",
			"message": fmt.Sprintf("%s is not found", DappfilePath),
		})
		os.Exit(16)
	}

	conf, err, warns := config.LoadDappfile(DappfilePath)
	if err != nil {
		fprintResponse(os.Stderr, map[string]string{
			"error":   "bad_dappfile",
			"warning": strings.Join(warns, "\n"),
			"message": fmt.Sprintf("Bad dappfile %s: %s", DappfilePath, err),
		})
		os.Exit(16)
	}

	serializedConfig, err := yaml.Marshal(yaml.MetaConfig{ImplicitDoc: false, Value: &conf})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot dump dappConfig yaml data: %s\n", err)
		os.Exit(1)
	}

	fprintResponse(os.Stdout, map[string]string{"dappConfig": string(serializedConfig), "warning": strings.Join(warns, "\n")})
}
