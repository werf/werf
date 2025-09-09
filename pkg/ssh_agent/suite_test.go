package ssh_agent

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSSHAgent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ssh-agent suite")
}
