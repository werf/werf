package host_cleaning_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHostCleaning(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Host Cleaning Suite")
}
