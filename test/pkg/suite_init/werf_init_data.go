package suite_init

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
)

type WerfInitData struct{}

func NewWerfInitData(tmpDirData *TmpDirData) *WerfInitData {
	data := &WerfInitData{}
	BeforeEach(func() {
		homeDir := filepath.Join(tmpDirData.TmpDir, "home")
		Expect(os.MkdirAll(homeDir, os.ModePerm)).To(Succeed())
		Expect(werf.Init(tmpDirData.TmpDir, homeDir)).To(Succeed())
	})
	return data
}
