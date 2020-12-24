package suite_init

import (
	"os"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/werf/werf/integration/utils"
)

type TmpDirData struct {
	TmpDir      string
	TestDirPath string
}

func (data *TmpDirData) Setup() bool {
	return SetupTmpDir(&data.TmpDir, &data.TestDirPath)
}

func SetupTmpDir(tmpDir, testDirPath *string) bool {
	ginkgo.BeforeEach(func() {
		*tmpDir = utils.GetTempDir()
		*testDirPath = *tmpDir
	})

	ginkgo.AfterEach(func() {
		err := os.RemoveAll(*tmpDir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	return true
}
