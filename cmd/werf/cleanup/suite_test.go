package cleanup

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdCleanup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Cleanup Suite")
}
