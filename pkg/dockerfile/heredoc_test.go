package dockerfile

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

var _ = Describe("MapToCorrectHeredocCmd", func() {
	DescribeTable("returns correct heredoc command",
		func(ctx SpecContext, cmdLine []string, files []instructions.ShellInlineFile, expectedCmd string, expectedPrependShell bool) {
			shellDependantCmdLine := instructions.ShellDependantCmdLine{CmdLine: cmdLine, Files: files}

			cmd, prependShell := MapToCorrectHeredocCmd(shellDependantCmdLine)

			Expect(cmd).To(Equal(expectedCmd))
			Expect(prependShell).To(Equal(expectedPrependShell))
		},

		Entry("no heredoc files – returns original command",
			[]string{"echo hello"},
			nil,
			"echo hello",
			true,
		),

		Entry("single heredoc file without chomp – keeps trailing newlines",
			[]string{"echo hello"},
			[]instructions.ShellInlineFile{
				{
					Name:  "EOF",
					Data:  "line1\nline2\n",
					Chomp: false,
				},
			},
			"echo hello\nline1\nline2\nEOF",
			true,
		),

		Entry("single heredoc file with chomp – trims trailing \\r\\n",
			[]string{"echo hello"},
			[]instructions.ShellInlineFile{
				{
					Name:  "EOF",
					Data:  "line1\r\nline2\r\n\r\n",
					Chomp: true,
				},
			},
			"echo hello\nline1\r\nline2EOF",
			true,
		),

		Entry("multiple heredoc files – concatenates in order",
			[]string{"echo hello"},
			[]instructions.ShellInlineFile{
				{
					Name:  "EOF1",
					Data:  "data1\n",
					Chomp: false,
				},
				{
					Name:  "EOF2",
					Data:  "data2\n",
					Chomp: false,
				},
			},
			strings.Join([]string{"echo hello", "data1\nEOF1", "data2\nEOF2"}, "\n"),
			true,
		),
	)
})
