package log_sanitize

import "regexp"

const DockerRateLimitSanitizedMessage = "You have reached your pull rate limit (credentials hidden)."

var DockerRateLimitCredsRegexp = regexp.MustCompile(
	`(?i)you have reached your pull rate limit as\s+'[^']+':.*?(?:\.|$)`,
)

func SanitizeDockerRateLimit(input string) string {
	return DockerRateLimitCredsRegexp.ReplaceAllString(
		input,
		DockerRateLimitSanitizedMessage,
	)
}
