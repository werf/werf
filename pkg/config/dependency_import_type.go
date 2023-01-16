package config

type DependencyImportType string

const (
	ImageNameImport   DependencyImportType = "ImageName"
	ImageTagImport    DependencyImportType = "ImageTag"
	ImageRepoImport   DependencyImportType = "ImageRepo"
	ImageIDImport     DependencyImportType = "ImageID"
	ImageDigestImport DependencyImportType = "ImageDigest"
)
