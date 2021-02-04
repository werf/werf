package helm

const (
	TrackTerminationModeAnnoName = "werf.io/track-termination-mode"

	FailModeAnnoName                  = "werf.io/fail-mode"
	FailuresAllowedPerReplicaAnnoName = "werf.io/failures-allowed-per-replica"

	LogRegexAnnoName      = "werf.io/log-regex"
	LogRegexForAnnoPrefix = "werf.io/log-regex-for-"

	SkipLogsAnnoName              = "werf.io/skip-logs"
	SkipLogsForContainersAnnoName = "werf.io/skip-logs-for-containers"
	ShowLogsOnlyForContainers     = "werf.io/show-logs-only-for-containers"
	ShowLogsUntilAnnoName         = "werf.io/show-logs-until"

	ShowEventsAnnoName = "werf.io/show-service-messages"

	ReplicasOnCreationAnnoName = "werf.io/replicas-on-creation"
)
