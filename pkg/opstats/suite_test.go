package opstats

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOpstats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Opstats Suite")
}
