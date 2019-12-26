package docs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"mvdan.cc/xurls"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/templates"
)

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("{{ header }} Options\n\n```shell\n")
		buf.WriteString(templates.FlagsUsages(flags))
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("{{ header }} Options inherited from parent commands\n\n```shell\n")
		buf.WriteString(templates.FlagsUsages(parentFlags))
		buf.WriteString("```\n\n")
	}
	return nil
}

func printEnvironments(buf *bytes.Buffer, cmd *cobra.Command) error {
	environments, ok := cmd.Annotations[common.CmdEnvAnno]
	if !ok {
		return nil
	}

	buf.WriteString("{{ header }} Environments\n\n```shell\n")
	buf.WriteString(environments)
	buf.WriteString("\n```\n\n")

	return nil
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	buf.WriteString(`{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
`)

	buf.WriteString(replaceLinks(long) + "\n\n")

	if cmd.Runnable() {
		buf.WriteString("{{ header }} Syntax\n\n")
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", templates.UsageLine(cmd)))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("{{ header }} Examples\n\n")
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", cmd.Example))
	}

	if err := printEnvironments(buf, cmd); err != nil {
		return err
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

func replaceLinks(s string) string {
	links := xurls.Relaxed.FindAllString(s, -1)
	for _, link := range links {
		s = strings.Replace(s, link, fmt.Sprintf("[%[1]s](%[1]s)", link), -1)
	}

	return s
}

// GenMarkdownTree will generate a markdown page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenMarkdownTree(cmd *cobra.Command, dir string) error {
	identity := func(s string) string { return s }
	emptyStr := func(s string) string { return "" }
	return GenMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

// GenMarkdownTreeCustom is the the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := cmd.CommandPath()
	basename = strings.Replace(basename, " ", "_", -1)
	basename = strings.Replace(basename, "-", "_", -1)
	basename = basename + ".md"

	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}
	if err := GenMarkdownCustom(cmd, f); err != nil {
		return err
	}
	return nil
}
