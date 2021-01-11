package suite_init

import (
	"os"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type WerfBinaryData struct {
	WerfBinPath string
}

func NewWerfBinaryData(synchronizedSuiteCallbacksData *SynchronizedSuiteCallbacksData) *WerfBinaryData {
	data := &WerfBinaryData{}
	synchronizedSuiteCallbacksData.SetSynchronizedBeforeSuiteNode1FuncWithReturnValue(ComputeWerfBinPath)
	synchronizedSuiteCallbacksData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(computedPath []byte) {
		data.WerfBinPath = string(computedPath)
	})
	synchronizedSuiteCallbacksData.AppendSynchronizedAfterSuiteAllNodesFunc(gexec.CleanupBuildArtifacts)
	return data
}

func ComputeWerfBinPath() []byte {
	werfBinPath := os.Getenv("WERF_TEST_BINARY_PATH")
	if werfBinPath == "" {
		var err error
		werfBinPath, err = gexec.Build("github.com/werf/werf/cmd/werf")
		Ω(err).ShouldNot(HaveOccurred())
	}

	return []byte(werfBinPath)
}
