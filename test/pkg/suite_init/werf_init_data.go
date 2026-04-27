package suite_init

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
)

type WerfInitData struct{}

func NewWerfInitData(tmpDirData *TmpDirData) *WerfInitData {
	data := &WerfInitData{}
	BeforeEach(func() {
		homeDir, err := os.MkdirTemp("", "werf-test-home-")
		Expect(err).To(Succeed())
		DeferCleanup(func() {
			Expect(os.RemoveAll(homeDir)).To(Succeed())
		})
		Expect(werf.Init(tmpDirData.TmpDir, homeDir)).To(Succeed())
	})
	return data
}
