package suite_init

import (
	"os"
	"runtime"

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
	synchronizedSuiteCallbacksData.AppendSynchronizedAfterSuiteNode1Func(gexec.CleanupBuildArtifacts)
	return data
}

func ComputeWerfBinPath() []byte {
	werfBinPath := os.Getenv("WERF_TEST_BINARY_PATH")
	if werfBinPath == "" {
		var err error

		// TODO: get rid of these hardcoded build instructions?
		if runtime.GOOS == "linux" {
			werfBinPath, err = gexec.BuildWithEnvironment("github.com/werf/werf/cmd/werf", []string{"CGO_ENABLED=1"}, "-compiler", "gc", "-ldflags", "-linkmode external -extldflags=-static", "-tags", "dfrunsecurity dfrunnetwork dfrunmount dfssh containers_image_openpgp osusergo exclude_graphdriver_devicemapper netgo no_devmapper static_build cni")
		} else {
			werfBinPath, err = gexec.BuildWithEnvironment("github.com/werf/werf/cmd/werf", nil, "-compiler", "gc", "-tags", "dfrunsecurity dfrunnetwork dfrunmount dfssh containers_image_openpgp")
		}
		Ω(err).ShouldNot(HaveOccurred())
	}

	return []byte(werfBinPath)
}
