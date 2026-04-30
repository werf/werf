package image

const (
	WerfLabelPrefix                        = "werf"
	WerfLabel                              = "werf"
	WerfVersionLabel                       = "werf-version"
	WerfStageContentDigestLabel            = "werf-stage-content-digest"
	WerfImageContextDigestLabel            = "werf-image-context-digest"
	WerfProjectRepoCommitLabel             = "werf-project-repo-commit"
	WerfImportSourceStageIDLabelPrefix     = "werf-import-source-stage-id-"
	WerfImportSourceExternalImagePrefix    = "external-image"
	WerfDependencySourceStageIDLabelPrefix = "werf-dependency-stage-id-"
	WerfBaseImageIDLabel                   = "werf.io/base-image-id"
	WerfParentStageID                      = "werf.io/parent-stage-id"

	WerfCustomTagMetadataStageIDLabel = "stage-id"
	WerfCustomTagMetadataTag          = "tag"

	WerfMountTmpDirLabel          = "werf-mount-type-tmp-dir"
	WerfMountBuildDirLabel        = "werf-mount-type-build-dir"
	WerfMountCustomDirLabelPrefix = "werf-mount-type-custom-dir-"

	BuildCacheVersion = "1.2"

	StageContainerNamePrefix        = "werf.build."
	ImportServerContainerNamePrefix = "import-server-"
	AssemblingContainerNamePrefix   = "werf.stapel."
)
