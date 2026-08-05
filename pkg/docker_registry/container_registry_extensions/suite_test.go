package container_registry_extensions
package container_registry_extensions

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestContainerRegistryExtensions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Container Registry Extensions Suite")
}
