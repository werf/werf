package image

const (
	WerfLabel                     = "werf"
	WerfVersionLabel              = "werf-version"
	WerfCacheVersionLabel         = "werf-cache-version"
	WerfImageLabel                = "werf-image"
	WerfDevLabel                  = "werf-dev"
	WerfDockerImageName           = "werf-docker-image-name"
	WerfStageDigestLabel          = "werf-stage-digest"
	WerfStageContentDigestLabel   = "werf-stage-content-digest"
	WerfProjectRepoCommitLabel    = "werf-project-repo-commit"
	WerfImportChecksumLabelPrefix = "werf-import-checksum-"

	WerfImportMetadataChecksumLabel       = "checksum"
	WerfImportMetadataSourceImageIDLabel  = "source-image-id"
	WerfImportMetadataImportSourceIDLabel = "import-source-id"

	WerfMountTmpDirLabel          = "werf-mount-type-tmp-dir"
	WerfMountBuildDirLabel        = "werf-mount-type-build-dir"
	WerfMountCustomDirLabelPrefix = "werf-mount-type-custom-dir-"

	BuildCacheVersion = "1.2"

	StageContainerNamePrefix = "werf.build."
)
