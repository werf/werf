package suite_init

import (
	"github.com/onsi/ginkgo"
	"github.com/prashantv/gostub"
)

type StubsData struct {
	Stubs *gostub.Stubs
}

func (data *StubsData) Setup() bool {
	data.Stubs = gostub.New()
	return SetupStubs(data.Stubs)
}

func SetupStubs(stubs *gostub.Stubs) bool {
	return ginkgo.AfterEach(func() {
		stubs.Reset()
	})
}
