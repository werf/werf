package opstats

import (
	"bytes"
	"context"
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
