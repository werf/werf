package opstats

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/werf/logboek"
)

func logSummaryOutput(collector *Collector) string {
	var out bytes.Buffer
	logger := logboek.NewLogger(&out, &out)
	logger.Streams().DisableStyle()
	LogSummary(logboek.NewContext(context.Background(), logger), collector, "build time", time.Second)
	return out.String()
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
