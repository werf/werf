package cleanup

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/pkg/cleaning"
)

func setupWhitelist(cmdData *cmdDataType, cmd *cobra.Command) {
	const name = "whitelist"

	cmd.Flags().StringVarP(&cmdData.Whitelist, name, "", os.Getenv("WERF_WHITELIST"), "Set file path to whitelist (default $WERF_WHITELIST)")
}

func parseWhitelist(filename string) (cleaning.Whitelist, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open whitelist file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to stat whitelist file: %w", err)
	}

	lineRegexp, err := regexp.Compile("^[a-z0-9]{56}-[0-9]{13}$") //nolint
	if err != nil {
		return nil, fmt.Errorf("unable to compile regexp: %w", err)
	}

	const lineWidthSize = 71

	whitelist := cleaning.NewWhitelistWithSize(int(stat.Size() / lineWidthSize))

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		tag := scanner.Text()

		switch {
		case lineRegexp.MatchString(tag):
			whitelist.Add(tag)
		case tag == "":
			continue
		default:
			return nil, fmt.Errorf("unable to parse tag %q", tag)
		}
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("scanner error: %w", scanner.Err())
	}

	return whitelist, nil
}
