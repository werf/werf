package suite_init

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/werf"
)

type WerfInitData struct{}

func NewWerfInitData(tmpDirData *TmpDirData) *WerfInitData {
	data := &WerfInitData{}
	BeforeEach(func() {
		Expect(werf.Init(tmpDirData.TmpDir, "")).To(Succeed())
	})
	return data
}
