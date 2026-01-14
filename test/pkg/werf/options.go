package werf

import "time"

type CommonOptions struct {
	ShouldFail bool
	ExtraArgs  []string

	CancelOnOutput        string
	CancelOnOutputTimeout time.Duration
}

type BuildOptions struct {
	CommonOptions
}

type WithReportOptions struct {
	CommonOptions
}

type ConvergeOptions struct {
	CommonOptions
}

type BundlePublishOptions struct {
	CommonOptions
}

type BundleApplyOptions struct {
	CommonOptions
}

type ExportOptions struct {
	CommonOptions
}

type CiEnvOptions struct {
	CommonOptions
}

type HostCleanupOptions struct {
	CommonOptions
}

type KubeRunOptions struct {
	CommonOptions
	Command []string
	Image   string
}

type StagesCopyOptions struct {
	CommonOptions
}

type KubeCtlOptions struct {
	CommonOptions
}
