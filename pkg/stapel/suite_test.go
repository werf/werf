package stapel

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAI_Stapel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stapel Suite")
}
