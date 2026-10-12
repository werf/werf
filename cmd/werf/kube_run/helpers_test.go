package kube_run

import (
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func clearListEnv() {
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "WERF_") {
			ginkgo.GinkgoT().Setenv(key, "")
			gomega.Expect(os.Unsetenv(key)).To(gomega.Succeed())
		}
	}
}
