package schemas

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSchemas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schemas Suite")
}
