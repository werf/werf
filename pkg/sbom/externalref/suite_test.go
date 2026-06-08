package externalref

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExternalRef(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "externalref suite")
}
