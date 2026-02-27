package manager

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStorageManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Manager Suite")
}
