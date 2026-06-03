package common

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAI_CmdCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Common Suite")
}
