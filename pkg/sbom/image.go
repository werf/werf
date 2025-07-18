package sbom

import "fmt"

func ImageName(name string) string {
	return fmt.Sprintf("%s-sbom", name)
}
