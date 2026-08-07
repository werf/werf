package stapel

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStapel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stapel Suite")
}
