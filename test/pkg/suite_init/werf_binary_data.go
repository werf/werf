package suite_init

import (
	"context"
	"os"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/go_task"
)

type WerfBinaryData struct {
	WerfBinPath string
}

func NewWerfBinaryData(synchronizedSuiteCallbacksData *SynchronizedSuiteCallbacksData) *WerfBinaryData {
	data := &WerfBinaryData{}
	synchronizedSuiteCallbacksData.SetSynchronizedBeforeSuiteNode1FuncWithReturnValue(ComputeWerfBinPath)
	synchronizedSuiteCallbacksData.AppendSynchronizedBeforeSuiteAllNodesFunc(func(_ context.Context, computedPath []byte) {
		data.WerfBinPath = string(computedPath)
	})
	synchronizedSuiteCallbacksData.AppendSynchronizedAfterSuiteNode1Func(func(_ context.Context) {
		gexec.CleanupBuildArtifacts()
	})
	return data
}

func ComputeWerfBinPath() []byte {
	werfBinPath := os.Getenv("WERF_TEST_BINARY_PATH")
	if werfBinPath == "" {
		ctx := context.TODO()
		werfBinPath = buildWerfDevBinary(ctx)
	}
	return []byte(werfBinPath)
}

func buildWerfDevBinary(ctx context.Context) string {
	basePath, err := utils.LookupRepoAbsPath(ctx)
	Expect(err).ShouldNot(HaveOccurred())

	buildOpts := go_task.BuildTaskOpts{
		AcceptPromptsAutomatically: true,
	}
	if e := os.Getenv("WERF_TEST_ENABLE_RACE_DETECTOR"); e != "" {
		buildOpts.RaceDetectorEnabled = true
	}
	werfBinPath, err := go_task.NewTaskfile("Taskfile.dist.yaml", basePath).BuildDevTask(ctx, buildOpts)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(werfBinPath).ShouldNot(BeEmpty())

	return werfBinPath
}
