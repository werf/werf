package gost

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGost(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gost Suite")
}
