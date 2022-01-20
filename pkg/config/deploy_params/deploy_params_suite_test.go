package deploy_params_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDeployParams(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DeployParams Suite")
}
