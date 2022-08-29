package bundles

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "deploy/bundles suite")
}
