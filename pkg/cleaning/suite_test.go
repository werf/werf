package cleaning

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test_Cleaning(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cleaning Suite")
}
