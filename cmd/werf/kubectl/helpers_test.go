package kubectl

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/kubectl/pkg/cmd"
)

func TestMain(m *testing.M) {
	if os.Getenv("WERF_TEST_KUBECTL_CONFIG") == "1" {
		command := NewCmd(context.Background())
		if err := command.ParseFlags(os.Args[2:]); err != nil {
			panic(err)
		}
		if err := json.NewEncoder(os.Stdout).Encode([]string{
			command.Flag("context").Value.String(),
			command.Flag("insecure-skip-tls-verify").Value.String(),
		}); err != nil {
			panic(err)
		}
		os.Exit(0)
	}
	if os.Getenv("WERF_TEST_KUBECTL_COMMAND") == "1" {
		if strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-") {
			if err := json.NewEncoder(os.Stdout).Encode(os.Args[1:]); err != nil {
				os.Exit(1)
			}
		} else {
			NewCmd(context.Background())
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func copyPlugin(path string) {
	ginkgo.GinkgoHelper()
	src, err := os.Open(os.Args[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer func() { gomega.Expect(src.Close()).To(gomega.Succeed()) }()
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer func() { gomega.Expect(dst.Close()).To(gomega.Succeed()) }()
	_, err = io.Copy(dst, src)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

var _ cmd.PluginHandler = (*fakePluginHandler)(nil)

type fakePluginHandler struct {
	installed []string
	executed  string
	execArgs  []string
}

func (h *fakePluginHandler) Lookup(filename string) (string, bool) {
	if slices.Contains(h.installed, filename) {
		return "/usr/local/bin/kubectl-" + filename, true
	}

	return "", false
}

func (h *fakePluginHandler) Execute(path string, args, _ []string) error {
	h.executed, h.execArgs = path, args
	return nil
}
