package opstats

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/onsi/gomega"

	"github.com/werf/logboek"
)

func logSummaryOutput(collector *Collector) string {
	var out bytes.Buffer
	logger := logboek.NewLogger(&out, &out)
	logger.Streams().DisableStyle()
	LogSummary(logboek.NewContext(context.Background(), logger), collector, "build time", time.Second)
	return out.String()
}

// summaryLineWidths returns the width of every rendered line as a terminal shows
// it: the logger used here has styling disabled, so a line is as wide as its rune
// count, and an escape sequence would mean it is not.
func summaryLineWidths(output string) []int {
	gomega.Expect(output).NotTo(gomega.ContainSubstring("\x1b"))

	var widths []int
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		widths = append(widths, utf8.RuneCountInString(line))
	}
	return widths
}

// blockingCloser holds Close until it is released, so a test can observe the state of the
// collector while the stream is still being closed.
type blockingCloser struct {
	io.Reader
	entered chan struct{}
	release chan struct{}
}

var _ io.ReadCloser = (*blockingCloser)(nil)

func (c *blockingCloser) Close() error {
	close(c.entered)
	<-c.release
	return nil
}
