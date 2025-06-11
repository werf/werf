package cleanup

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/werf/werf/v2/pkg/cleaning"
)

func setupKeeplist(cmdData *cmdDataType, cmd *cobra.Command) {
	const name = "keep-list"

	cmd.Flags().StringVarP(&cmdData.KeepList, name, "", os.Getenv("WERF_KEEP_LIST"), "Set path to keep list file (default $WERF_KEEP_LIST)")
}

func parseKeepList(filename string) (cleaning.KeepList, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open keep list file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to stat keep list file: %w", err)
	}

	lineRegexp, err := regexp.Compile("^[a-z0-9]{56}-[0-9]{13}$") //nolint
	if err != nil {
		return nil, fmt.Errorf("unable to compile regexp: %w", err)
	}

	const lineWidthSize = 71

	keepList := cleaning.NewKeepListWithSize(int(stat.Size() / lineWidthSize))

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		tag := scanner.Text()

		switch {
		case lineRegexp.MatchString(tag):
			keepList.Add(tag)
		case tag == "":
			continue
		default:
			return nil, fmt.Errorf("unable to parse tag %q", tag)
		}
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("scanner error: %w", scanner.Err())
	}

	return keepList, nil
}
