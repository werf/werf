package sbom

const (
	ScannerImage = "ghcr.io/anchore/syft:v1.22.0"
)

type ScanOptions struct {
	Image      string
	PullPolicy PullPolicy

	// TODO (zaytsev): give an able to use multiple source locations
	// Commands        []string      // one or more commands to invoke for the image rootfs or ContextDir locations
	// ContextDir      []string      // one or more "source" directory locations
	// MergeStrategy   MergeStrategy // how to merge the outputs of multiple scans

	SBOMOutput string // where to save SBOM scanner output outside of the image (i.e., the local filesystem)
	// ImageSBOMOutput string // where to save SBOM scanner output in the image
}
