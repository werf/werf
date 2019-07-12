package docs

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
)

var CmdData struct {
	dest        string
	readmePath  string
	splitReadme bool
}
var CommonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "docs",
		DisableFlagsInUseLine: true,
		Short:                 "Generate documentation as markdown",
		Hidden:                true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&CommonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if CmdData.splitReadme {
				if err := SplitReadme(); err != nil {
					return err
				}
			} else {
				if err := GenMarkdownTree(cmd.Root(), CmdData.dest); err != nil {
					return err
				}
			}

			return nil
		},
	}

	common.SetupLogOptions(&CommonCmdData, cmd)

	f := cmd.Flags()
	f.StringVar(&CmdData.dest, "dir", "./", "directory to which documentation is written")
	f.StringVar(&CmdData.readmePath, "readme", "README.md", "path to README.md")
	f.BoolVar(&CmdData.splitReadme, "split-readme", false, "split README.md by top headers")

	return cmd
}

func SplitReadme() error {
	file, err := os.Open(CmdData.readmePath)
	if err != nil {
		return err
	}

	defer file.Close()

	currentTitle := "main"
	partialsData := map[string][]string{}
	partialsData[currentTitle] = []string{}

	splitRegexp := regexp.MustCompile(`^# (.*)`)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := splitRegexp.FindStringSubmatch(line); len(matches) != 0 {
			title := matches[1]
			partialsData[title] = []string{}
			currentTitle = title
		} else {
			partialsData[currentTitle] = append(partialsData[currentTitle], line)
		}
	}

	for header, data := range partialsData {
		basename := strings.ToLower(header)
		basename = strings.Replace(basename, " ", "_", -1)
		basename = strings.Replace(basename, "-", "_", -1)
		basename = basename + ".md"

		filename := filepath.Join(CmdData.dest, basename)
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.WriteString(f, strings.Join(data, "\n")); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}
