package cleanup_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/werf/werf/v2/pkg/build/cleanup"
)

var _ = DescribeTable("Promise",
	func(spyCalls []*spyCall, expectedCallsCount int, useForget bool) {
		var composedFunc cleanup.Func

		defer func() {
			assertCalls(spyCalls, lo.Ternary(useForget, 0, expectedCallsCount))
			composedFunc()
			assertCalls(spyCalls, expectedCallsCount)
		}()

		p := cleanup.NewPromise()
		defer p.Give()

		for _, spy := range spyCalls {
			p.Add(spy.Callback)
		}

		composedFunc = lo.TernaryF(useForget, p.Forget, func() cleanup.Func {
			return cleanup.NoOp
		})
	},
	Entry(
		"should call single deferred callback",
		[]*spyCall{
			newSpyCall(false),
		},
		1,
		false,
	),
	Entry(
		"should call two deferred callbacks (as composed function) even if first one panics",
		[]*spyCall{
			newSpyCall(true),
			newSpyCall(false),
		},
		2,
		false,
	),
	Entry(
		"should forget deferred call and return composed function which could be called manually",
		[]*spyCall{
			newSpyCall(false),
			newSpyCall(true),
		},
		2,
		true,
	),
)

func assertCalls(spyCalls []*spyCall, expectedCallsCount int) {
	actualCallsCount := lo.SumBy(spyCalls, func(spy *spyCall) int {
		return spy.CallsCount()
	})
	Expect(actualCallsCount).To(Equal(expectedCallsCount))
}

type spyCall struct {
	callsCount int
	usePanic   bool
}

func (s *spyCall) Callback() {
	s.callsCount++
	if s.usePanic {
		panic("callback panic")
	}
}

func (s *spyCall) CallsCount() int {
	return s.callsCount
}

func newSpyCall(usePanic bool) *spyCall {
	return &spyCall{
		usePanic: usePanic,
	}
}
