package suite_init

import (
	"github.com/onsi/ginkgo"
	"github.com/prashantv/gostub"
)

type StubsData struct {
	Stubs *gostub.Stubs
}

func NewStubsData() *StubsData {
	data := &StubsData{
		Stubs: gostub.New(),
	}
	SetupStubs(data.Stubs)
	return data
}

func SetupStubs(stubs *gostub.Stubs) bool {
	return ginkgo.AfterEach(func() {
		stubs.Reset()
	})
}
