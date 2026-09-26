package build

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/config"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/giterminism_manager"
	"github.com/werf/werf/v3/pkg/storage/manager"
)

var _ = ginkgo.DescribeTable("Remote preparation limit", func(parallel bool, limit int64, expected int) {
	conveyor := NewConveyor(
		&config.WerfConfig{Meta: &config.Meta{Project: "preparation"}},
		&giterminism_manager.Manager{}, ".", ginkgo.GinkgoT().TempDir(),
		&container_backend.DockerServerBackend{}, &manager.StorageManager{},
		ConveyorOptions{Parallel: parallel, ParallelTasksLimit: limit},
	)
	gomega.Expect(conveyor.imagesTree.RemoteGitTasksLimit).To(gomega.Equal(expected))
},
	ginkgo.Entry("disabled", false, int64(8), 1),
	ginkgo.Entry("negative limit", true, int64(-1), 4),
	ginkgo.Entry("unlimited build", true, int64(0), 4),
	ginkgo.Entry("one worker", true, int64(1), 1),
	ginkgo.Entry("lower limit", true, int64(2), 2),
	ginkgo.Entry("bounded default", true, int64(8), 4),
)
