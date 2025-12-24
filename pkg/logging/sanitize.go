package logging

import (
	"io"
	"regexp"
)

type SanitizeWriter struct {
	w  io.Writer
	re *regexp.Regexp
}

var dockerRateLimitCredsRe = regexp.MustCompile(
	`(?i)you have reached your pull rate limit as\s+'[^']+':.*?(?:\.|$)`,
)

func (s *SanitizeWriter) Write(p []byte) (int, error) {
	out := s.re.ReplaceAllString(
		string(p),
		"You have reached your pull rate limit (credentials hidden).",
	)
	return s.w.Write([]byte(out))
}

func WrapIfNeeded(w io.Writer) io.Writer {
	return &SanitizeWriter{
		w:  w,
		re: dockerRateLimitCredsRe,
	}
}
