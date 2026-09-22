package container_backend

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("LegacyBaseImage", func() {
	It("synchronizes concurrent info updates", func() {
		baseImage := newLegacyBaseImage("image", nil)
		infos := []*image.Info{{ID: "first"}, {ID: "second"}}

		var waitGroup sync.WaitGroup
		for _, info := range infos {
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				for range 1000 {
					baseImage.SetInfo(info)
					_ = baseImage.GetInfo()
					_ = baseImage.IsExistsLocally()
					baseImage.SetStageDesc(&image.StageDesc{Info: info})
					_ = baseImage.GetStageDesc().Info
				}
			}()
		}
		waitGroup.Wait()

		baseImage.SetStageDesc(&image.StageDesc{Info: infos[0]})
		assert.Same(GinkgoT(), infos[0], baseImage.GetInfo())
	})
})
