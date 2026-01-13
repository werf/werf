package logging

import (
	"io"
	"regexp"

	"github.com/werf/werf/v2/pkg/log_sanitize"
)

type SanitizeWriter struct {
	w  io.Writer
	re *regexp.Regexp
}

func (s *SanitizeWriter) Write(p []byte) (int, error) {
	out := log_sanitize.SanitizeDockerRateLimit(string(p))
	return s.w.Write([]byte(out))
}

func WrapInSanitizeWriter(w io.Writer) io.Writer {
	return &SanitizeWriter{
		w:  w,
		re: log_sanitize.DockerRateLimitCredsRegexp,
	}
}
